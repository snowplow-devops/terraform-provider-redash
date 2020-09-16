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
			"redash_group":                        resourceRedashGroup(),
			"redash_group_user_attachment":        resourceRedashGroupUserAttachment(),
			"redash_group_data_source_attachment": resourceRedashGroupDataSourceAttachment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"redash_data_source": dataSourceRedashDataSource(),
			"redash_user":        dataSourceRedashUser(),
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
