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

func userResourceData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceRedashUser().Schema, raw)
}

func TestResourceUser_createWithoutGroups(t *testing.T) {
	f, c := newFakeRedash(t)

	d := userResourceData(t, map[string]interface{}{"name": "Wile E. Coyote", "email": "wcoyote@acme.com"})

	if diags := resourceRedashUserCreate(context.Background(), d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if d.Id() != "1" {
		t.Fatalf("expected id 1, got %q", d.Id())
	}
	if got := f.requestsMatching("POST /api/users/"); len(got) != 0 {
		t.Errorf("expected no group update without groups, got %v", got)
	}

	if d.Get("name") != "Wile E. Coyote" || d.Get("email") != "wcoyote@acme.com" {
		t.Errorf("unexpected state: name=%q email=%q", d.Get("name"), d.Get("email"))
	}
	if got := d.Get("auth_type"); got != "password" {
		t.Errorf("expected auth_type password, got %q", got)
	}
	if got := d.Get("is_invitation_pending"); got != true {
		t.Errorf("expected is_invitation_pending true, got %v", got)
	}
	if got, want := d.Get("groups"), []interface{}{fakeDefaultGroupID}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups: got %v, want %v", got, want)
	}
}

func TestResourceUser_createWithGroups(t *testing.T) {
	f, c := newFakeRedash(t)

	d := userResourceData(t, map[string]interface{}{
		"name":   "Road Runner",
		"email":  "rrunner@acme.com",
		"groups": []interface{}{3, 4},
	})

	if diags := resourceRedashUserCreate(context.Background(), d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if got, want := f.user(1).Groups, []int{3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups in Redash: got %v, want %v", got, want)
	}
	if got, want := d.Get("groups"), []interface{}{3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups in state: got %v, want %v", got, want)
	}
}

func TestResourceUser_createGroupUpdateFails(t *testing.T) {
	f, c := newFakeRedash(t)
	f.failOn["POST /api/users/1"] = http.StatusBadRequest

	d := userResourceData(t, map[string]interface{}{
		"name":   "Road Runner",
		"email":  "rrunner@acme.com",
		"groups": []interface{}{3},
	})

	diags := resourceRedashUserCreate(context.Background(), d, c)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "400") {
		t.Fatalf("expected the group update error, got %v", diags)
	}
	// The user was created, so it stays in state for Terraform to taint.
	if d.Id() != "1" {
		t.Errorf("expected id 1 to be kept, got %q", d.Id())
	}
}

func TestResourceUser_updateAndDelete(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addUser(&fakeUser{Name: "Wile E. Coyote", Email: "wcoyote@acme.com"})

	d := userResourceData(t, map[string]interface{}{
		"name":   "Wile E. Coyote, Genius",
		"email":  "genius@acme.com",
		"groups": []interface{}{2, 5},
	})
	d.SetId("1")

	if diags := resourceRedashUserUpdate(ctx, d, c); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}
	u := f.user(1)
	if u.Name != "Wile E. Coyote, Genius" || u.Email != "genius@acme.com" || !reflect.DeepEqual(u.Groups, []int{2, 5}) {
		t.Errorf("unexpected user in Redash: %+v", u)
	}

	d.Set("groups", []interface{}{})
	if diags := resourceRedashUserUpdate(ctx, d, c); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}
	if got := f.user(1).Groups; len(got) != 0 {
		t.Errorf("expected groups to be cleared, got %v", got)
	}

	if diags := resourceRedashUserDelete(ctx, d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
	if !f.user(1).IsDisabled {
		t.Errorf("expected delete to disable the user in Redash")
	}
}

func TestResourceUser_import(t *testing.T) {
	f, c := newFakeRedash(t)
	f.addUser(&fakeUser{Name: "Wile E. Coyote", Email: "wcoyote@acme.com", IsEmailVerified: true})

	d := userResourceData(t, map[string]interface{}{})
	d.SetId("1")

	if diags := resourceRedashUserRead(context.Background(), d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Get("name") != "Wile E. Coyote" || d.Get("email") != "wcoyote@acme.com" || d.Get("is_email_verified") != true {
		t.Errorf("unexpected state: name=%q email=%q verified=%v", d.Get("name"), d.Get("email"), d.Get("is_email_verified"))
	}
}

func TestResourceUser_apiErrors(t *testing.T) {
	f, c := newFakeRedash(t)
	d := userResourceData(t, map[string]interface{}{"name": "Wile E. Coyote", "email": "wcoyote@acme.com"})

	f.failWith = http.StatusInternalServerError
	if diags := resourceRedashUserCreate(context.Background(), d, c); !diags.HasError() {
		t.Errorf("expected create to fail")
	}
	if d.Id() != "" {
		t.Errorf("expected no id after a failed create, got %q", d.Id())
	}
	f.failWith = 0

	assertAPIErrors(t, f, d, c, crudFuncs(resourceRedashUserRead, resourceRedashUserUpdate, resourceRedashUserDelete))
}

func TestDataSourceUser(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addUser(&fakeUser{Name: "Wile E. Coyote", Email: "wcoyote@acme.com"})
	f.addUser(&fakeUser{Name: "Road Runner", Email: "rrunner@acme.com"})

	d := schema.TestResourceDataRaw(t, dataSourceRedashUser().Schema, map[string]interface{}{"email": "rrunner@acme.com"})
	if diags := dataSourceRedashUserRead(ctx, d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != "2" {
		t.Errorf("expected id 2, got %q", d.Id())
	}

	d = schema.TestResourceDataRaw(t, dataSourceRedashUser().Schema, map[string]interface{}{"email": "nobody@acme.com"})
	diags := dataSourceRedashUserRead(ctx, d, c)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "no user found") {
		t.Errorf("expected a not found error, got %v", diags)
	}

	f.failWith = http.StatusInternalServerError
	d = schema.TestResourceDataRaw(t, dataSourceRedashUser().Schema, map[string]interface{}{"email": "rrunner@acme.com"})
	if diags := dataSourceRedashUserRead(ctx, d, c); !diags.HasError() {
		t.Errorf("expected an error")
	}
}
