package redash

import (
	"fmt"

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
				Type:     schema.TypeString,
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
			"view_only": {
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
			/*"options": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Schema: map[string]*schema.Schema{
						"warehouse": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dbname": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},*/
		},
		Read: dataSourceRedashDataSourceRead,
	}
}

func dataSourceRedashDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*redash.Client)

	id := d.Get("id").(int)

	data_source, err := c.GetDataSource(id)
	if err != nil {
		return err
	}

	d.Set("name", data_source.Name)

	//d.Set("options", "{\"test\": \"a thing\"}")

	fmt.Println(*data_source.Options)
	d.SetId(fmt.Sprint(data_source.ID))

	return nil
}
