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
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scheduled_queue_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"queue_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"paused": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"pause_reason": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"syntax": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"options": {
				Type:     schema.TypeMap,
				Optional: true,
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
