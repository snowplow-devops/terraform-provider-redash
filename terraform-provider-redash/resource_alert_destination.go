//
// Copyright (c) 2020-2026 Snowplow Analytics Ltd. All rights reserved.
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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

// redashSecretPlaceholder is what Redash returns in place of secret option
// values (webhook passwords, integration keys, etc.) when reading them back.
const redashSecretPlaceholder = "--------"

type alertDestinationPayload struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}

// alertDestinationFields holds the fields common to every destination type,
// with options flattened to strings to fit the Terraform schema.
type alertDestinationFields struct {
	ID      int
	Name    string
	Type    string
	Icon    string
	Options map[string]string
}

func resourceRedashAlertDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedashAlertDestinationCreate,
		ReadContext:   resourceRedashAlertDestinationRead,
		UpdateContext: resourceRedashAlertDestinationUpdate,
		DeleteContext: resourceRedashAlertDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: resourceRedashAlertDestinationCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"options": {
				Type:      schema.TypeMap,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"icon": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// alertDestinationTypes maps each destination type the client can parse to
// its typed struct. The client drops any option it doesn't declare on these
// structs, so they define which options can round-trip through Terraform.
var alertDestinationTypes = map[string]reflect.Type{
	"email":         reflect.TypeOf(redash.EmailDestination{}),
	"slack":         reflect.TypeOf(redash.SlackDestination{}),
	"webhook":       reflect.TypeOf(redash.WebhookDestination{}),
	"mattermost":    reflect.TypeOf(redash.MattermostDestination{}),
	"chatwork":      reflect.TypeOf(redash.ChatWorkDestination{}),
	"pagerduty":     reflect.TypeOf(redash.PagerDutyDestination{}),
	"hangouts_chat": reflect.TypeOf(redash.HangoutsChatDestination{}),
}

// alertDestinationOptionKeys returns the option keys supported for a
// destination type, or false if the type isn't one the client parses (in
// which case any options are passed through untouched).
func alertDestinationOptionKeys(destinationType string) (map[string]bool, bool) {
	t, ok := alertDestinationTypes[destinationType]
	if !ok {
		return nil, false
	}

	options, _ := t.FieldByName("Options")
	keys := map[string]bool{}
	for i := 0; i < options.Type.NumField(); i++ {
		tag := options.Type.Field(i).Tag.Get("json")
		keys[strings.Split(tag, ",")[0]] = true
	}
	return keys, true
}

func validateAlertDestinationOptions(destinationType string, options map[string]interface{}) error {
	allowed, ok := alertDestinationOptionKeys(destinationType)
	if !ok {
		return nil
	}

	var unsupported []string
	for k := range options {
		if !allowed[k] {
			unsupported = append(unsupported, k)
		}
	}
	if len(unsupported) == 0 {
		return nil
	}

	var supported []string
	for k := range allowed {
		supported = append(supported, k)
	}
	sort.Strings(unsupported)
	sort.Strings(supported)

	return fmt.Errorf("unsupported options for %q alert destination: %s (supported: %s)",
		destinationType, strings.Join(unsupported, ", "), strings.Join(supported, ", "))
}

func resourceRedashAlertDestinationCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// A value that's unknown until apply (e.g. taken from a resource that
	// hasn't been created yet) reads as empty here, so it passes validation.
	return validateAlertDestinationOptions(d.Get("type").(string), d.Get("options").(map[string]interface{}))
}

func alertDestinationPayloadFromResourceData(d *schema.ResourceData) ([]byte, error) {
	options := map[string]string{}
	for k, v := range d.Get("options").(map[string]interface{}) {
		options[k] = v.(string)
	}

	return json.Marshal(alertDestinationPayload{
		Name:    d.Get("name").(string),
		Type:    d.Get("type").(string),
		Options: options,
	})
}

// flattenAlertDestination converts whatever typed struct the client returned
// into the common fields. An unsupported destination type is not an error
// here: the client still returns a generic *redash.Destination for it.
func flattenAlertDestination(destination interface{}, err error) (*alertDestinationFields, error) {
	if err != nil && !errors.Is(err, redash.ErrUnsupportedDestinationType) {
		return nil, err
	}
	if destination == nil {
		return nil, fmt.Errorf("empty alert destination response")
	}

	raw, err := json.Marshal(destination)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		ID      int                    `json:"id"`
		Name    string                 `json:"name"`
		Type    string                 `json:"type"`
		Icon    string                 `json:"icon"`
		Options map[string]interface{} `json:"options"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}

	fields := &alertDestinationFields{
		ID:      parsed.ID,
		Name:    parsed.Name,
		Type:    parsed.Type,
		Icon:    parsed.Icon,
		Options: map[string]string{},
	}
	for k, v := range parsed.Options {
		if v == nil {
			continue
		}
		s := fmt.Sprint(v)
		if s == "" {
			continue
		}
		fields.Options[k] = s
	}

	return fields, nil
}

// mergeAlertDestinationOptions keeps the value from state for any option that
// Redash returns masked, so secrets don't show a diff on every plan. A masked
// option with no prior value (e.g. on import) is left out.
func mergeAlertDestinationOptions(remote map[string]string, prior map[string]interface{}) map[string]string {
	merged := map[string]string{}
	for k, v := range remote {
		if v == redashSecretPlaceholder {
			if p, ok := prior[k]; ok {
				merged[k] = p.(string)
			}
			continue
		}
		merged[k] = v
	}
	return merged
}

func isNotFoundError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "HTTP Response: 404")
}

func resourceRedashAlertDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	payload, err := alertDestinationPayloadFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	destination, err := flattenAlertDestination(c.CreateDestination(payload))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(destination.ID))

	return resourceRedashAlertDestinationRead(ctx, d, meta)
}

func resourceRedashAlertDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	destination, err := flattenAlertDestination(c.GetDestination(id))
	if isNotFoundError(err) {
		d.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", destination.Name)
	d.Set("type", destination.Type)
	d.Set("icon", destination.Icon)
	d.Set("options", mergeAlertDestinationOptions(destination.Options, d.Get("options").(map[string]interface{})))

	d.SetId(fmt.Sprint(destination.ID))

	return diags
}

func resourceRedashAlertDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	payload, err := alertDestinationPayloadFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = flattenAlertDestination(c.UpdateDestination(id, payload))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRedashAlertDestinationRead(ctx, d, meta)
}

func resourceRedashAlertDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*redash.Client)

	var diags diag.Diagnostics

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteDestination(id)
	if err != nil && !isNotFoundError(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
