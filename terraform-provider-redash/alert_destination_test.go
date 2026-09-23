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
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/snowplow-devops/redash-client-go/redash"
)

var fakeSecretOptions = map[string]bool{"password": true, "integration_key": true, "api_token": true}

var fakeIcons = map[string]string{"email": "fa-envelope", "slack": "fa-slack", "webhook": "fa-bolt"}

func (f *fakeRedash) addDestination(destination map[string]interface{}) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.nextDestinationID
	f.nextDestinationID++
	destination["id"] = id
	destination["icon"] = fakeIcons[destination["type"].(string)]
	f.destinations[id] = destination
	return id
}

func (f *fakeRedash) destination(id int) map[string]interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.destinations[id]
}

// handleDestinations fakes the destinations API. Like the real thing, it
// masks secret options when returning a single destination and omits
// options from the list endpoint.
func (f *fakeRedash) handleDestinations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/destinations")
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			f.listDestinations(w)
		case http.MethodPost:
			var destination map[string]interface{}
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &destination); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id := f.addDestination(destination)
			f.writeDestination(w, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	id, err := strconv.Atoi(strings.TrimPrefix(path, "/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if f.destination(id) == nil {
		http.Error(w, `{"message": "Not found"}`, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		f.writeDestination(w, id)
	case http.MethodPost:
		var update map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		for k, v := range update {
			f.destinations[id][k] = v
		}
		f.mu.Unlock()
		f.writeDestination(w, id)
	case http.MethodDelete:
		f.mu.Lock()
		delete(f.destinations, id)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *fakeRedash) listDestinations(w http.ResponseWriter) {
	f.mu.Lock()
	defer f.mu.Unlock()

	list := []map[string]interface{}{}
	for id := 1; id < f.nextDestinationID; id++ {
		if d, ok := f.destinations[id]; ok {
			list = append(list, map[string]interface{}{"id": d["id"], "name": d["name"], "type": d["type"], "icon": d["icon"]})
		}
	}
	_ = json.NewEncoder(w).Encode(list)
}

func (f *fakeRedash) writeDestination(w http.ResponseWriter, id int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	d := f.destinations[id]
	masked := map[string]interface{}{}
	for k, v := range d["options"].(map[string]interface{}) {
		if fakeSecretOptions[k] {
			v = redashSecretPlaceholder
		}
		masked[k] = v
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id": d["id"], "name": d["name"], "type": d["type"], "icon": d["icon"], "options": masked,
	})
}

func terraformResourceConfig(raw map[string]interface{}) *terraform.ResourceConfig {
	return terraform.NewResourceConfigRaw(raw)
}

func alertDestinationResourceData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceRedashAlertDestination().Schema, raw)
}

func stringMap(v interface{}) map[string]string {
	out := map[string]string{}
	for k, val := range v.(map[string]interface{}) {
		out[k] = val.(string)
	}
	return out
}

func TestResourceAlertDestination_lifecycle(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()

	d := alertDestinationResourceData(t, map[string]interface{}{
		"name": "Ops webhook",
		"type": "webhook",
		"options": map[string]interface{}{
			"url":      "https://hooks.example.com/redash",
			"username": "redash",
			"password": "s3cret",
		},
	})

	if diags := resourceRedashAlertDestinationCreate(ctx, d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	if d.Id() != "1" {
		t.Fatalf("expected id 1, got %q", d.Id())
	}
	if got := d.Get("icon"); got != "fa-bolt" {
		t.Errorf("expected icon fa-bolt, got %q", got)
	}
	wantOptions := map[string]string{
		"url":      "https://hooks.example.com/redash",
		"username": "redash",
		"password": "s3cret",
	}
	if got := stringMap(d.Get("options")); !reflect.DeepEqual(got, wantOptions) {
		t.Errorf("options after create: got %v, want %v", got, wantOptions)
	}
	if got := f.destination(1)["options"].(map[string]interface{})["password"]; got != "s3cret" {
		t.Errorf("expected the real password to be sent to Redash, got %v", got)
	}

	// A refresh should keep the secret from state rather than the mask.
	if diags := resourceRedashAlertDestinationRead(ctx, d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if got := d.Get("options.password"); got != "s3cret" {
		t.Errorf("expected password to survive refresh, got %q", got)
	}

	d.Set("name", "Ops webhook v2")
	d.Set("options", map[string]interface{}{
		"url":      "https://hooks.example.com/v2",
		"password": "n3w",
	})
	if diags := resourceRedashAlertDestinationUpdate(ctx, d, c); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}

	stored := f.destination(1)
	if stored["name"] != "Ops webhook v2" {
		t.Errorf("expected name to be updated in Redash, got %v", stored["name"])
	}
	storedOptions := stored["options"].(map[string]interface{})
	if storedOptions["url"] != "https://hooks.example.com/v2" || storedOptions["password"] != "n3w" {
		t.Errorf("expected options to be updated in Redash, got %v", storedOptions)
	}
	if _, ok := storedOptions["username"]; ok {
		t.Errorf("expected removed option to be dropped, got %v", storedOptions)
	}
	if got := d.Get("name"); got != "Ops webhook v2" {
		t.Errorf("expected name in state to be updated, got %q", got)
	}

	if diags := resourceRedashAlertDestinationDelete(ctx, d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
	if f.destination(1) != nil {
		t.Errorf("expected destination to be deleted from Redash")
	}
}

func TestResourceAlertDestination_readRemovesMissing(t *testing.T) {
	_, c := newFakeRedash(t)

	d := alertDestinationResourceData(t, map[string]interface{}{"name": "gone", "type": "email"})
	d.SetId("42")

	if diags := resourceRedashAlertDestinationRead(context.Background(), d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected a destination deleted outside Terraform to be removed from state, got id %q", d.Id())
	}
}

func TestResourceAlertDestination_deleteMissingIsNotAnError(t *testing.T) {
	_, c := newFakeRedash(t)

	d := alertDestinationResourceData(t, map[string]interface{}{"name": "gone", "type": "email"})
	d.SetId("42")

	if diags := resourceRedashAlertDestinationDelete(context.Background(), d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
}

func TestResourceAlertDestination_import(t *testing.T) {
	f, c := newFakeRedash(t)
	id := f.addDestination(map[string]interface{}{
		"name": "PagerDuty",
		"type": "pagerduty",
		"options": map[string]interface{}{
			"integration_key": "abc123",
			"description":     "Redash alert",
		},
	})

	d := alertDestinationResourceData(t, map[string]interface{}{})
	d.SetId(fmt.Sprint(id))

	if diags := resourceRedashAlertDestinationRead(context.Background(), d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}

	if got := d.Get("name"); got != "PagerDuty" {
		t.Errorf("expected name PagerDuty, got %q", got)
	}
	if got := d.Get("type"); got != "pagerduty" {
		t.Errorf("expected type pagerduty, got %q", got)
	}
	// The masked secret can't be recovered on import, so it's left out
	// rather than stored as the placeholder.
	want := map[string]string{"description": "Redash alert"}
	if got := stringMap(d.Get("options")); !reflect.DeepEqual(got, want) {
		t.Errorf("options after import: got %v, want %v", got, want)
	}
}

func TestResourceAlertDestination_unsupportedType(t *testing.T) {
	_, c := newFakeRedash(t)
	ctx := context.Background()

	d := alertDestinationResourceData(t, map[string]interface{}{
		"name":    "Legacy",
		"type":    "hipchat",
		"options": map[string]interface{}{"url": "https://hipchat.example.com", "room": "ops"},
	})

	if diags := resourceRedashAlertDestinationCreate(ctx, d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	want := map[string]string{"url": "https://hipchat.example.com", "room": "ops"}
	if got := stringMap(d.Get("options")); !reflect.DeepEqual(got, want) {
		t.Errorf("options for an unsupported type should pass through: got %v, want %v", got, want)
	}
}

func TestResourceAlertDestination_apiErrors(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.failWith = http.StatusInternalServerError

	d := alertDestinationResourceData(t, map[string]interface{}{
		"name":    "Email",
		"type":    "email",
		"options": map[string]interface{}{"addresses": "ops@example.com"},
	})

	if diags := resourceRedashAlertDestinationCreate(ctx, d, c); !diags.HasError() {
		t.Errorf("expected create to fail")
	}
	if d.Id() != "" {
		t.Errorf("expected no id after a failed create, got %q", d.Id())
	}

	d.SetId("1")
	for name, fn := range map[string]schema.ReadContextFunc{
		"read":   resourceRedashAlertDestinationRead,
		"update": resourceRedashAlertDestinationUpdate,
		"delete": resourceRedashAlertDestinationDelete,
	} {
		diags := fn(ctx, d, c)
		if !diags.HasError() {
			t.Errorf("expected %s to fail", name)
			continue
		}
		if !strings.Contains(diags[0].Summary, "500") {
			t.Errorf("expected %s error to include the status code, got %q", name, diags[0].Summary)
		}
	}

	d.SetId("not-a-number")
	if diags := resourceRedashAlertDestinationRead(ctx, d, c); !diags.HasError() {
		t.Errorf("expected read with a non-numeric id to fail")
	}
}

func TestAlertDestinationOptionKeys(t *testing.T) {
	for destinationType := range alertDestinationTypes {
		keys, ok := alertDestinationOptionKeys(destinationType)
		if !ok || len(keys) == 0 {
			t.Errorf("expected option keys for %q, got %v", destinationType, keys)
		}
	}

	keys, _ := alertDestinationOptionKeys("email")
	if want := map[string]bool{"addresses": true, "subject_template": true}; !reflect.DeepEqual(keys, want) {
		t.Errorf("email option keys: got %v, want %v", keys, want)
	}

	if _, ok := alertDestinationOptionKeys("hipchat"); ok {
		t.Errorf("expected no option keys for an unsupported type")
	}
}

func TestValidateAlertDestinationOptions(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		options map[string]interface{}
		wantErr string
	}{
		{"valid", "slack", map[string]interface{}{"url": "u", "channel": "#ops"}, ""},
		{"empty", "email", map[string]interface{}{}, ""},
		{"unsupported type passes through", "hipchat", map[string]interface{}{"anything": "x"}, ""},
		{"unknown key", "email", map[string]interface{}{"addresses": "a", "cc": "b", "bcc": "c"},
			`unsupported options for "email" alert destination: bcc, cc (supported: addresses, subject_template)`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAlertDestinationOptions(tc.typ, tc.options)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Errorf("unexpected error: %s", err)
			case tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr):
				t.Errorf("got error %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestFlattenAlertDestination(t *testing.T) {
	if _, err := flattenAlertDestination(nil, errors.New("boom")); err == nil || err.Error() != "boom" {
		t.Errorf("expected the client error to be returned, got %v", err)
	}

	if _, err := flattenAlertDestination(nil, nil); err == nil {
		t.Errorf("expected an error for an empty response")
	}

	unsupported := &redash.Destination{
		ID:   7,
		Name: "Custom",
		Type: "custom",
		Options: map[string]interface{}{
			"enabled": true,
			"retries": 3,
			"blank":   "",
			"missing": nil,
		},
	}
	fields, err := flattenAlertDestination(unsupported, fmt.Errorf("%w: %q", redash.ErrUnsupportedDestinationType, "custom"))
	if err != nil {
		t.Fatalf("an unsupported type should not be an error: %s", err)
	}
	if fields.ID != 7 || fields.Name != "Custom" || fields.Type != "custom" {
		t.Errorf("unexpected fields: %+v", fields)
	}
	if want := map[string]string{"enabled": "true", "retries": "3"}; !reflect.DeepEqual(fields.Options, want) {
		t.Errorf("options: got %v, want %v", fields.Options, want)
	}
}

func TestMergeAlertDestinationOptions(t *testing.T) {
	remote := map[string]string{"url": "u", "password": redashSecretPlaceholder, "api_token": redashSecretPlaceholder}
	prior := map[string]interface{}{"url": "old", "password": "s3cret"}

	want := map[string]string{"url": "u", "password": "s3cret"}
	if got := mergeAlertDestinationOptions(remote, prior); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResourceAlertDestination_customizeDiff(t *testing.T) {
	p := Provider()
	res := p.ResourcesMap["redash_alert_destination"]

	cfg := map[string]interface{}{
		"name":    "Email",
		"type":    "email",
		"options": map[string]interface{}{"addresses": "ops@example.com", "cc": "x"},
	}
	_, err := res.Diff(context.Background(), nil, terraformResourceConfig(cfg), nil)
	if err == nil || !strings.Contains(err.Error(), `unsupported options for "email"`) {
		t.Errorf("expected plan to fail on an unsupported option, got %v", err)
	}

	cfg["options"] = map[string]interface{}{"addresses": "ops@example.com"}
	if _, err := res.Diff(context.Background(), nil, terraformResourceConfig(cfg), nil); err != nil {
		t.Errorf("unexpected plan error: %s", err)
	}
}

// unknownValue marks a config value as unknown until apply, e.g. one taken
// from a resource that hasn't been created yet. It matches the SDK's
// hcl2shim.UnknownVariableValue, which is internal and can't be imported.
const unknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

func TestResourceAlertDestination_customizeDiffSkipsUnknownValues(t *testing.T) {
	res := Provider().ResourcesMap["redash_alert_destination"]

	cases := map[string]map[string]interface{}{
		// The unsupported "cc" option can't be checked without knowing the type.
		"unknown type": {
			"name":    "Email",
			"type":    unknownValue,
			"options": map[string]interface{}{"addresses": "ops@example.com", "cc": "x"},
		},
		"unknown options": {
			"name":    "Email",
			"type":    "email",
			"options": unknownValue,
		},
		// One unknown value makes the SDK treat the whole map as unknown.
		"unknown option value": {
			"name":    "Email",
			"type":    "email",
			"options": map[string]interface{}{"addresses": unknownValue, "cc": "x"},
		},
	}

	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			diff, err := res.Diff(context.Background(), nil, terraformResourceConfig(cfg), nil)
			if err != nil {
				t.Fatalf("expected an unknown value not to fail the plan, got %s", err)
			}
			if diff == nil {
				t.Fatalf("expected a diff for a new resource")
			}
		})
	}
}

func dataSourceAlertDestinationData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, dataSourceRedashAlertDestination().Schema, raw)
}

func TestDataSourceAlertDestination(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addDestination(map[string]interface{}{"name": "Email", "type": "email", "options": map[string]interface{}{}})
	f.addDestination(map[string]interface{}{"name": "Slack", "type": "slack", "options": map[string]interface{}{}})
	f.addDestination(map[string]interface{}{"name": "Dupe", "type": "webhook", "options": map[string]interface{}{}})
	f.addDestination(map[string]interface{}{"name": "Dupe", "type": "webhook", "options": map[string]interface{}{}})

	t.Run("by id", func(t *testing.T) {
		d := dataSourceAlertDestinationData(t, map[string]interface{}{"id": 2})
		if diags := dataSourceRedashAlertDestinationRead(ctx, d, c); diags.HasError() {
			t.Fatalf("read: %v", diags)
		}
		if d.Id() != "2" || d.Get("name") != "Slack" || d.Get("type") != "slack" || d.Get("icon") != "fa-slack" {
			t.Errorf("unexpected result: id=%q name=%q type=%q icon=%q", d.Id(), d.Get("name"), d.Get("type"), d.Get("icon"))
		}
	})

	t.Run("by name", func(t *testing.T) {
		d := dataSourceAlertDestinationData(t, map[string]interface{}{"name": "Email"})
		if diags := dataSourceRedashAlertDestinationRead(ctx, d, c); diags.HasError() {
			t.Fatalf("read: %v", diags)
		}
		if d.Id() != "1" || d.Get("id") != 1 || d.Get("type") != "email" {
			t.Errorf("unexpected result: id=%q type=%q", d.Id(), d.Get("type"))
		}
	})

	errorCases := []struct {
		name    string
		raw     map[string]interface{}
		wantErr string
	}{
		{"missing id", map[string]interface{}{"id": 99}, "no alert destination found with id 99"},
		{"missing name", map[string]interface{}{"name": "Nope"}, `no alert destination found with name "Nope"`},
		{"duplicate name", map[string]interface{}{"name": "Dupe"}, `2 alert destinations found with name "Dupe", use id instead`},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			d := dataSourceAlertDestinationData(t, tc.raw)
			diags := dataSourceRedashAlertDestinationRead(ctx, d, c)
			if !diags.HasError() || diags[0].Summary != tc.wantErr {
				t.Errorf("got %v, want error %q", diags, tc.wantErr)
			}
		})
	}

	t.Run("api error", func(t *testing.T) {
		f.failWith = http.StatusBadGateway
		defer func() { f.failWith = 0 }()

		d := dataSourceAlertDestinationData(t, map[string]interface{}{"id": 1})
		if diags := dataSourceRedashAlertDestinationRead(ctx, d, c); !diags.HasError() {
			t.Errorf("expected an error")
		}
	})
}

func TestDataSourceAlertDestination_requiresExactlyOneOf(t *testing.T) {
	ds := Provider().DataSourcesMap["redash_alert_destination"]

	for name, cfg := range map[string]map[string]interface{}{
		"neither": {},
		"both":    {"id": 1, "name": "Email"},
	} {
		t.Run(name, func(t *testing.T) {
			if diags := ds.Validate(terraformResourceConfig(cfg)); !diags.HasError() {
				t.Errorf("expected validation to fail")
			}
		})
	}
}
