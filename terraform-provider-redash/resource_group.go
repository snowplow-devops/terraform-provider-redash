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
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func resourceRedashGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashGroupCreate,
		ReadContext:   resourceRedashGroupRead,
		UpdateContext: resourceRedashGroupUpdate,
		DeleteContext: resourceRedashGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRedashGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	payload := redash.GroupCreatePayload{
		Name: d.Get("name").(string),
	}

	group, err := c.CreateGroup(&payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(group.ID))

	resourceRedashGroupRead(ctx, d, meta)

	return diags
}

func resourceRedashGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := c.GetGroup(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", &group.Name)
	d.Set("type", &group.Type)
	d.Set("permissions", &group.Permissions)

	d.SetId(fmt.Sprint(group.ID))

	return diags
}

func resourceRedashGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	payload := redash.Group{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	_, err = c.UpdateGroup(id, &payload)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRedashGroupRead(ctx, d, meta)
}

func resourceRedashGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteGroup(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
