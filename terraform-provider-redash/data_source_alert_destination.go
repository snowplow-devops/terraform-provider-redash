//
// Copyright (c) 2020-2026 Snowplow Analytics Ltd. All rights reserved.
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
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func dataSourceRedashAlertDestination() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"icon": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		ReadContext: dataSourceRedashAlertDestinationRead,
	}
}

func dataSourceRedashAlertDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	destinations, err := c.GetDestinations()
	if err != nil {
		return diag.FromErr(err)
	}

	id, byID := d.GetOk("id")
	name := d.Get("name").(string)

	var matches []redash.Destination
	for _, destination := range *destinations {
		if (byID && destination.ID == id.(int)) || (!byID && destination.Name == name) {
			matches = append(matches, destination)
		}
	}

	switch {
	case len(matches) == 0 && byID:
		return diag.Errorf("no alert destination found with id %d", id.(int))
	case len(matches) == 0:
		return diag.Errorf("no alert destination found with name %q", name)
	case len(matches) > 1:
		return diag.Errorf("%d alert destinations found with name %q, use id instead", len(matches), name)
	}

	destination := matches[0]

	d.Set("id", destination.ID)
	d.Set("name", destination.Name)
	d.Set("type", destination.Type)
	d.Set("icon", destination.Icon)

	d.SetId(fmt.Sprint(destination.ID))

	return diags
}
