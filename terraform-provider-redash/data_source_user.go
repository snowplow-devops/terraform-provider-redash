package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func dataSourceRedashUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ReadContext: dataSourceRedashUserRead,
	}
}

func dataSourceRedashUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	email := d.Get("email").(string)
	user, err := c.GetUserByEmail(email)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", user.Name)

	d.SetId(fmt.Sprint(user.ID))

	return diags
}
