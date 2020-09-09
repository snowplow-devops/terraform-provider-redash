package redash

import (
	"context"
	"fmt"
	"strconv"

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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
		ReadContext: dataSourceRedashDataSourceRead,
	}
}

func dataSourceRedashDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id := d.Get("id").(int)

	data_source, err := c.GetDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", data_source.Name)
	d.Set("scheduled_queue_name", data_source.ScheduledQueueName)
	d.Set("pause_reason", data_source.PauseReason)
	d.Set("queue_name", data_source.QueueName)
	d.Set("syntax", data_source.Syntax)
	d.Set("paused", data_source.Paused)
	d.Set("type", data_source.Type)
	options := flattenOptionsData(data_source)
	if err := d.Set("options", options); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(data_source.Id))

	return diags
}

func flattenOptionsData(data_source *redash.DataSource) map[string]interface{} {

	if data_source != nil {
		switch {
		case data_source.Type == "redshift":
			return map[string]interface{}{
				"port":   strconv.Itoa(data_source.Options.Port),
				"host":   data_source.Options.Host,
				"user":   data_source.Options.User,
				"dbname": data_source.Options.Dbname,
			}
		case data_source.Type == "bigquery":
			return map[string]interface{}{
				"project_id":       data_source.Options.ProjectId,
				"use_standard_sql": strconv.FormatBool(data_source.Options.UseStandardSQL),
				"load_schema":      strconv.FormatBool(data_source.Options.LoadSchema),
			}
		case data_source.Type == "snowflake":
			return map[string]interface{}{
				"warehouse": data_source.Options.Warehouse,
				"account":   data_source.Options.Account,
				"user":      data_source.Options.User,
				"database":  data_source.Options.Database,
			}
		}

	}

	return map[string]interface{}{}
}
