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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func resourceRedashGroupDataSourceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashGroupDataSourceAttachmentCreate,
		ReadContext:   resourceRedashGroupDataSourceAttachmentRead,
		DeleteContext: resourceRedashGroupDataSourceAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"data_source_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRedashGroupDataSourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	dataSourceID := d.Get("data_source_id").(int)

	err := c.GroupAddDataSource(groupID, dataSourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%d-%d", groupID, dataSourceID)))

	return diags
}

func resourceRedashGroupDataSourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	dataSourceID := d.Get("data_source_id").(int)

	dataSource, err := c.GetDataSource(dataSourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := dataSource.Groups[groupID]; ok {
		return diags
	}

	d.SetId("")

	return diags
}

func resourceRedashGroupDataSourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	dataSourceID := d.Get("data_source_id").(int)

	err := c.GroupRemoveDataSource(groupID, dataSourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
