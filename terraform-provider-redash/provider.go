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
		},
		ResourcesMap: map[string]*schema.Resource{
			"redash_data_source":                  resourceRedashDataSource(),
			"redash_user":                         resourceRedashUser(),
			"redash_group":                        resourceRedashGroup(),
			"redash_group_data_source_attachment": resourceRedashGroupDataSourceAttachment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"redash_data_source": dataSourceRedashDataSource(),
			"redash_user":        dataSourceRedashUser(),
			"redash_group":       dataSourceRedashGroup(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	APIKey := d.Get("api_key").(string)
	RedashURI := d.Get("redash_uri").(string)

	var diags diag.Diagnostics

	c, err := redash.NewClient(&redash.Config{RedashURI: RedashURI, APIKey: APIKey})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Redash API Client Error",
			Detail:   err.Error(),
		})
	}

	return c, diags
}
