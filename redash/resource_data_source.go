package redash

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func resourceRedashDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashDataSourceCreate,
		ReadContext:   resourceRedashDataSourceRead,
		UpdateContext: resourceRedashDataSourceUpdate,
		DeleteContext: resourceRedashDataSourceDelete,
		Schema: map[string]*schema.Schema{
			// "last_updated": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// 	Computed: true,
			// },
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
				Optional: true,
				Default:  0,
			},
			"pause_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"syntax": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"warehouse": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"account": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "--------" {
									return true
								}

								return false

							},
						},
						"user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"host": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dbname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"use_standard_sql": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"json_key_file": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return true
							},
						},
						"load_schema": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceRedashDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	options := redash.Options{}
	mapstructure.Decode(d.Get("options").([]interface{})[0], &options)

	payload := redash.DataSource{
		Name:               d.Get("name").(string),
		Type:               d.Get("type").(string),
		ScheduledQueueName: d.Get("scheduled_queue_name").(string),
		PauseReason:        d.Get("pause_reason").(string),
		QueueName:          d.Get("queue_name").(string),
		Syntax:             d.Get("syntax").(string),
		Paused:             d.Get("paused").(int),
		Options:            options,
	}

	data_source, err := c.CreateDataSource(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(data_source.Id))

	resourceRedashDataSourceRead(ctx, d, meta)

	return diags
}

func resourceRedashDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	data_source, err := c.GetDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", &data_source.Name)
	d.Set("scheduled_queue_name", &data_source.ScheduledQueueName)
	d.Set("pause_reason", &data_source.PauseReason)
	d.Set("queue_name", &data_source.QueueName)
	d.Set("syntax", &data_source.Syntax)
	d.Set("paused", &data_source.Paused)
	d.Set("type", &data_source.Type)

	options := flattenOptions(data_source)
	if err := d.Set("options", []interface{}{options}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(data_source.Id))

	return diags
}

func resourceRedashDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	options := redash.Options{}
	mapstructure.Decode(d.Get("options").([]interface{})[0], &options)

	payload := redash.DataSource{
		Name:               d.Get("name").(string),
		Type:               d.Get("type").(string),
		ScheduledQueueName: d.Get("scheduled_queue_name").(string),
		PauseReason:        d.Get("pause_reason").(string),
		QueueName:          d.Get("queue_name").(string),
		Syntax:             d.Get("syntax").(string),
		Paused:             d.Get("paused").(int),
		Options:            options,
	}

	_, err = c.UpdateDataSource(id, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRedashDataSourceRead(ctx, d, meta)
}

func resourceRedashDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteDataSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func flattenOptions(data_source *redash.DataSource) map[string]interface{} {

	if data_source != nil {
		switch {
		case data_source.Type == "redshift":
			return map[string]interface{}{
				"port":     data_source.Options.Port,
				"host":     data_source.Options.Host,
				"user":     data_source.Options.User,
				"password": data_source.Options.Password,
				"dbname":   data_source.Options.Dbname,
			}
		case data_source.Type == "bigquery":
			return map[string]interface{}{
				"project_id":       data_source.Options.ProjectId,
				"use_standard_sql": strconv.FormatBool(data_source.Options.UseStandardSQL),
				"json_key_file":    data_source.Options.JSONKeyFile,
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
