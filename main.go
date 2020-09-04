package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/snowplow-devops/terraform-provider-redash/redash"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return redash.Provider()
		},
	})
}
