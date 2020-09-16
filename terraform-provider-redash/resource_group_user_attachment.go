package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

func resourceRedashGroupUserAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashGroupUserAttachmentCreate,
		ReadContext:   resourceRedashGroupUserAttachmentRead,
		DeleteContext: resourceRedashGroupUserAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRedashGroupUserAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	userID := d.Get("user_id").(int)

	err := c.GroupAddUser(groupID, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%d-%d", groupID, userID)))

	return diags
}

func resourceRedashGroupUserAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	userID := d.Get("user_id").(int)

	user, err := c.GetUser(userID)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, gid := range user.Groups {
		if gid == groupID {
			return diags
		}
	}

	d.SetId("")

	return diags
}

func resourceRedashGroupUserAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	userID := d.Get("user_id").(int)

	err := c.GroupRemoveUser(groupID, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
