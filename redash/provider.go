package redash

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/snowplow-devops/redash-client-go/redash"
)

// Provider
func Provider() terraform.ResourceProvider {
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
			//"redash_data_source": resourceRedashDataSource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"redash_data_source": dataSourceRedashDataSource(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	api_key := d.Get("api_key").(string)
	hostname := d.Get("hostname").(string)

	if (api_key != "") && (hostname != "") {
		c, err := redash.NewClient(&hostname, &api_key)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
