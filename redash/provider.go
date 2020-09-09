package redash

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

type Config struct {
	RedashURL string
	APIKey    string
}

// Provider
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
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("REDASH_URL", ""),
				Description: "Redash API Key",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"redash_data_source": resourceRedashDataSource(),
			"redash_group":       resourceRedashGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"redash_data_source": dataSourceRedashDataSource(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	api_key := d.Get("api_key").(string)
	hostname := d.Get("hostname").(string)

	var diags diag.Diagnostics

	if (api_key == "") || (hostname == "") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Redash client",
			Detail:   "Missing REDASH_API_KEY or REDASH_URL environment variables.",
		})
	}
	c := redash.NewClient(&redash.Config{RedashURL: hostname, APIKey: api_key})

	return c, diags
}
