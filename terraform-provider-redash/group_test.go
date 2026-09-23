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
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// crudFuncs lists a resource's CRUD functions (other than create) so error
// handling can be tested for each in turn.
func crudFuncs(read, update, del schema.ReadContextFunc) map[string]schema.ReadContextFunc {
	funcs := map[string]schema.ReadContextFunc{"read": read, "delete": del}
	if update != nil {
		funcs["update"] = update
	}
	return funcs
}

// assertAPIErrors checks that each function surfaces a failing API call and
// rejects a non-numeric ID.
func assertAPIErrors(t *testing.T, f *fakeRedash, d *schema.ResourceData, meta interface{}, funcs map[string]schema.ReadContextFunc) {
	t.Helper()
	ctx := context.Background()

	f.failWith = http.StatusInternalServerError
	defer func() { f.failWith = 0 }()

	for name, fn := range funcs {
		d.SetId("1")
		diags := fn(ctx, d, meta)
		if !diags.HasError() {
			t.Errorf("expected %s to fail", name)
		} else if !strings.Contains(diags[0].Summary, "500") {
			t.Errorf("expected %s error to include the status code, got %q", name, diags[0].Summary)
		}

		d.SetId("not-a-number")
		if diags := fn(ctx, d, meta); !diags.HasError() {
			t.Errorf("expected %s with a non-numeric id to fail", name)
		}
	}
}

func TestResourceGroup_lifecycle(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()

	d := schema.TestResourceDataRaw(t, resourceRedashGroup().Schema, map[string]interface{}{"name": "Analysts"})

	if diags := resourceRedashGroupCreate(ctx, d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if d.Id() != "1" {
		t.Fatalf("expected id 1, got %q", d.Id())
	}
	if got := f.group(1); got == nil || got.Name != "Analysts" {
		t.Fatalf("expected group to be created in Redash, got %+v", got)
	}
	if got := d.Get("type"); got != "regular" {
		t.Errorf("expected type regular, got %q", got)
	}
	if got, want := d.Get("permissions"), []interface{}{"create_query", "view_query"}; !reflect.DeepEqual(got, want) {
		t.Errorf("permissions: got %v, want %v", got, want)
	}

	d.Set("name", "Data Analysts")
	if diags := resourceRedashGroupUpdate(ctx, d, c); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}
	if got := f.group(1).Name; got != "Data Analysts" {
		t.Errorf("expected name to be updated in Redash, got %q", got)
	}
	if got := d.Get("name"); got != "Data Analysts" {
		t.Errorf("expected name in state to be updated, got %q", got)
	}

	if diags := resourceRedashGroupDelete(ctx, d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
	if f.group(1) != nil {
		t.Errorf("expected group to be deleted from Redash")
	}
}

func TestResourceGroup_import(t *testing.T) {
	f, c := newFakeRedash(t)
	g := f.addGroup("Admins")

	d := schema.TestResourceDataRaw(t, resourceRedashGroup().Schema, map[string]interface{}{})
	d.SetId("1")

	if diags := resourceRedashGroupRead(context.Background(), d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Get("name") != g.Name || d.Get("type") != g.Type {
		t.Errorf("unexpected state: name=%q type=%q", d.Get("name"), d.Get("type"))
	}
}

func TestResourceGroup_apiErrors(t *testing.T) {
	f, c := newFakeRedash(t)
	d := schema.TestResourceDataRaw(t, resourceRedashGroup().Schema, map[string]interface{}{"name": "Analysts"})

	f.failWith = http.StatusInternalServerError
	if diags := resourceRedashGroupCreate(context.Background(), d, c); !diags.HasError() {
		t.Errorf("expected create to fail")
	}
	if d.Id() != "" {
		t.Errorf("expected no id after a failed create, got %q", d.Id())
	}
	f.failWith = 0

	assertAPIErrors(t, f, d, c, crudFuncs(resourceRedashGroupRead, resourceRedashGroupUpdate, resourceRedashGroupDelete))
}

func TestDataSourceGroup(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addGroup("Admins")
	f.addGroup("Analysts")

	d := schema.TestResourceDataRaw(t, dataSourceRedashGroup().Schema, map[string]interface{}{"id": 2})
	if diags := dataSourceRedashGroupRead(ctx, d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != "2" || d.Get("name") != "Analysts" {
		t.Errorf("unexpected result: id=%q name=%q", d.Id(), d.Get("name"))
	}

	d = schema.TestResourceDataRaw(t, dataSourceRedashGroup().Schema, map[string]interface{}{"id": 99})
	diags := dataSourceRedashGroupRead(ctx, d, c)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "404") {
		t.Errorf("expected a 404 error for a missing group, got %v", diags)
	}
}
