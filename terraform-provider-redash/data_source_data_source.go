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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func dataSourceRedashDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scheduled_queue_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"queue_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"paused": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pause_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"syntax": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
		ReadContext: dataSourceRedashDataSourceRead,
	}
}

func dataSourceRedashDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id := d.Get("id").(int)

	dataSource, err := c.GetDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", dataSource.Name)
	d.Set("scheduled_queue_name", dataSource.ScheduledQueueName)
	d.Set("pause_reason", dataSource.PauseReason)
	d.Set("queue_name", dataSource.QueueName)
	d.Set("syntax", dataSource.Syntax)
	d.Set("paused", dataSource.Paused)
	d.Set("type", dataSource.Type)
	d.Set("options", dataSource.Options)

	d.SetId(fmt.Sprint(dataSource.ID))

	return diags
}
