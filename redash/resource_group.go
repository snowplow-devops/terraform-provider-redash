package redash

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
				Optional: true,
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

	payload := redash.Group{
		Name: d.Get("name").(string),
		//Permissions: []d.Get("permissions").([]string),
	}

	group, err := c.CreateGroup(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(group.Id))

	//resourceRedashGroupRead(ctx, d, meta)

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

	d.SetId(fmt.Sprint(group.Id))

	return diags
}

func resourceRedashGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	payload := redash.Group{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Permissions: d.Get("permissions").([]string),
	}

	_, err = c.UpdateGroup(id, payload)
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
