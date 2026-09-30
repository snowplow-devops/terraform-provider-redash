//
// Copyright (c) 2020 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
//
package main

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

// Provider function
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("REDASH_API_KEY", ""),
				Description: "Redash API Key",
			},
			"redash_uri": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("REDASH_URL", ""),
				Description: "Redash API Key",
			},
			"http_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Extra HTTP headers sent on every Redash API request.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"redash_data_source":                  resourceRedashDataSource(),
			"redash_user":                         resourceRedashUser(),
			"redash_group":                        resourceRedashGroup(),
			"redash_group_data_source_attachment": resourceRedashGroupDataSourceAttachment(),
			"redash_alert_destination":            resourceRedashAlertDestination(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"redash_data_source":       dataSourceRedashDataSource(),
			"redash_user":              dataSourceRedashUser(),
			"redash_group":             dataSourceRedashGroup(),
			"redash_alert_destination": dataSourceRedashAlertDestination(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	APIKey := d.Get("api_key").(string)
	RedashURI := d.Get("redash_uri").(string)

	var diags diag.Diagnostics

	cfg := &redash.Config{RedashURI: RedashURI, APIKey: APIKey}
	if headers := httpHeaders(d); len(headers) > 0 {
		cfg.HTTPClient = &http.Client{Transport: &headerTransport{headers: headers}}
	}

	c, err := redash.NewClient(cfg)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Redash API Client Error",
			Detail:   err.Error(),
		})
	}

	return c, diags
}

// httpHeaders reads the optional http_headers provider argument.
// The Terraform SDK stores map values as map[string]interface{}, so this
// converts them to strings for the HTTP transport. An unset or empty
// argument returns nil, and the Redash client then uses http.DefaultClient.
func httpHeaders(d *schema.ResourceData) map[string]string {
	raw, ok := d.GetOk("http_headers")
	if !ok {
		return nil
	}

	in := raw.(map[string]interface{})
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v.(string)
	}
	return out
}

// headerTransport sets configured headers on each request. RoundTrip must not
// modify the incoming request, so it clones it first. The Redash client sets
// Authorization before it calls http.Client.Do, which then runs RoundTrip, so
// that header is already on the request and is left in place.
type headerTransport struct {
	headers map[string]string
	base    http.RoundTripper
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
