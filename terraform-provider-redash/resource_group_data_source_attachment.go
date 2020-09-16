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
