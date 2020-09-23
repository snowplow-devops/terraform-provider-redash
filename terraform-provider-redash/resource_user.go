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

func resourceRedashUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashUserCreate,
		ReadContext:   resourceRedashUserRead,
		UpdateContext: resourceRedashUserUpdate,
		DeleteContext: resourceRedashUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"auth_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_image_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_invitation_pending": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disabled_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_email_verified": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"active_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRedashUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	createPayload := redash.UserCreatePayload{
		Name:  d.Get("name").(string),
		Email: d.Get("email").(string),
	}

	user, err := c.CreateUser(&createPayload)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(user.ID))

	groupIds := []int{}

	if raw, ok := d.GetOk("groups"); ok {

		list := raw.([]interface{})

		for _, v := range list {
			groupID, ok := v.(int)
			if !ok {
				return diag.FromErr(fmt.Errorf("invalid group_id found"))
			}
			groupIds = append(groupIds, groupID)
		}
	}

	if len(groupIds) > 0 {
		updatePayload := redash.UserUpdatePayload{
			Name:   d.Get("name").(string),
			Email:  d.Get("email").(string),
			Groups: groupIds,
		}
		_, err := c.UpdateUser(user.ID, &updatePayload)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceRedashUserRead(ctx, d, meta)

	return diags
}

func resourceRedashUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := c.GetUser(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", &user.Name)
	d.Set("email", &user.Email)
	d.Set("groups", &user.Groups)
	d.Set("auth_type", &user.AuthType)
	d.Set("is_disabled", &user.IsDisabled)
	d.Set("updated_at", &user.UpdatedAt)
	d.Set("profile_image_url", &user.ProfileImageURL)
	d.Set("is_invitation_pending", &user.IsInvitationPending)
	d.Set("created_at", &user.CreatedAt)
	d.Set("disabled_at", &user.DisabledAt)
	d.Set("is_email_verified", &user.IsEmailVerified)
	d.Set("active_at", &user.ActiveAt)

	d.SetId(fmt.Sprint(user.ID))

	return diags
}

func resourceRedashUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	groupIds := []int{}

	if raw, ok := d.GetOk("groups"); ok {

		list := raw.([]interface{})

		for _, v := range list {
			groupID, ok := v.(int)
			if !ok {
				return diag.FromErr(fmt.Errorf("invalid group_id found"))
			}
			groupIds = append(groupIds, groupID)
		}
	}

	updatePayload := redash.UserUpdatePayload{
		Name:   d.Get("name").(string),
		Email:  d.Get("email").(string),
		Groups: groupIds,
	}

	_, err = c.UpdateUser(id, &updatePayload)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRedashUserRead(ctx, d, meta)
}

func resourceRedashUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)
	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DisableUser(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
