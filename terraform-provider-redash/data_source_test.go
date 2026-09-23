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

func dataSourceResourceData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceRedashDataSource().Schema, raw)
}

func pgDataSourceConfig(name string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"type": "pg",
		"options": []interface{}{map[string]interface{}{
			"host":     "db.acme.com",
			"port":     5432,
			"user":     "wcoyote",
			"password": "eth3LbeRt",
			"dbname":   "products",
		}},
	}
}

func TestResourceDataSource_lifecycle(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()

	d := dataSourceResourceData(t, pgDataSourceConfig("ACME Products"))

	if diags := resourceRedashDataSourceCreate(ctx, d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if d.Id() != "1" {
		t.Fatalf("expected id 1, got %q", d.Id())
	}

	// Options that aren't part of the pg configuration schema are dropped by
	// the client before they reach Redash.
	stored := f.dataSource(1)
	wantOptions := map[string]interface{}{
		"host":     "db.acme.com",
		"port":     float64(5432),
		"user":     "wcoyote",
		"password": "eth3LbeRt",
		"dbname":   "products",
		"sslmode":  "",
	}
	if !reflect.DeepEqual(stored.Options, wantOptions) {
		t.Errorf("options in Redash: got %v, want %v", stored.Options, wantOptions)
	}
	if stored.Name != "ACME Products" || stored.Type != "pg" {
		t.Errorf("unexpected data source in Redash: %+v", stored)
	}

	if d.Get("name") != "ACME Products" || d.Get("type") != "pg" || d.Get("syntax") != "sql" {
		t.Errorf("unexpected state: name=%q type=%q syntax=%q", d.Get("name"), d.Get("type"), d.Get("syntax"))
	}

	d.Set("name", "ACME Products (replica)")
	if diags := resourceRedashDataSourceUpdate(ctx, d, c); diags.HasError() {
		t.Fatalf("update: %v", diags)
	}
	if got := f.dataSource(1).Name; got != "ACME Products (replica)" {
		t.Errorf("expected name to be updated in Redash, got %q", got)
	}
	if got := d.Get("name"); got != "ACME Products (replica)" {
		t.Errorf("expected name in state to be updated, got %q", got)
	}

	if diags := resourceRedashDataSourceDelete(ctx, d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
	if f.dataSource(1) != nil {
		t.Errorf("expected data source to be deleted from Redash")
	}
}

func TestResourceDataSource_convertsOptionsForRedash(t *testing.T) {
	f, c := newFakeRedash(t)

	// mongodb isn't in the fake's type list, so options reach Redash as the
	// provider sends them.
	d := dataSourceResourceData(t, map[string]interface{}{
		"name": "ACME Mongo",
		"type": "mongodb",
		"options": []interface{}{map[string]interface{}{
			"connection_string": "mongodb://db.acme.com",
			"db_name":           "products",
			"ssh_tunnel": []interface{}{map[string]interface{}{
				"ssh_host":     "bastion.acme.com",
				"ssh_port":     22,
				"ssh_username": "tunnel",
			}},
		}},
	})

	if diags := resourceRedashDataSourceCreate(context.Background(), d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}

	options := f.dataSource(1).Options
	if options["connectionString"] != "mongodb://db.acme.com" || options["dbName"] != "products" {
		t.Errorf("expected camelCase option names, got connectionString=%v dbName=%v", options["connectionString"], options["dbName"])
	}
	if _, ok := options["db_name"]; ok {
		t.Errorf("expected db_name to be renamed, got %v", options)
	}
	wantTunnel := map[string]interface{}{"ssh_host": "bastion.acme.com", "ssh_port": float64(22), "ssh_username": "tunnel"}
	if !reflect.DeepEqual(options["ssh_tunnel"], wantTunnel) {
		t.Errorf("ssh_tunnel: got %v, want %v", options["ssh_tunnel"], wantTunnel)
	}
}

func TestResourceDataSource_apiErrors(t *testing.T) {
	f, c := newFakeRedash(t)
	d := dataSourceResourceData(t, pgDataSourceConfig("ACME Products"))

	f.failWith = http.StatusInternalServerError
	if diags := resourceRedashDataSourceCreate(context.Background(), d, c); !diags.HasError() {
		t.Errorf("expected create to fail")
	}
	if d.Id() != "" {
		t.Errorf("expected no id after a failed create, got %q", d.Id())
	}

	// With every request failing, update stops when the client fetches the
	// data source types to validate options, so also fail just the update.
	f.failWith = 0
	f.addDataSource(&fakeDataSource{Name: "ACME Products", Type: "pg"})
	f.failOn["POST /api/data_sources/1"] = http.StatusInternalServerError
	d.SetId("1")
	if diags := resourceRedashDataSourceUpdate(context.Background(), d, c); !diags.HasError() {
		t.Errorf("expected update to fail")
	}

	assertAPIErrors(t, f, d, c, crudFuncs(resourceRedashDataSourceRead, resourceRedashDataSourceUpdate, resourceRedashDataSourceDelete))
}

func TestConvertOptions(t *testing.T) {
	cases := []struct {
		name     string
		toFormat string
		in       map[string]interface{}
		want     map[string]interface{}
	}{
		{
			name:     "to redash",
			toFormat: "redash",
			in:       map[string]interface{}{"db_name": "d", "use_standard_sql": true, "host": "h"},
			want:     map[string]interface{}{"dbName": "d", "useStandardSql": true, "host": "h"},
		},
		{
			name:     "to terraform",
			toFormat: "terraform",
			in:       map[string]interface{}{"dbName": "d", "totalMBytesProcessedLimit": 10, "host": "h"},
			want:     map[string]interface{}{"db_name": "d", "total_mbytes_processed_limit": 10, "host": "h"},
		},
		{
			name:     "ssh tunnel is unwrapped from its list",
			toFormat: "redash",
			in:       map[string]interface{}{"ssh_tunnel": []interface{}{map[string]interface{}{"ssh_host": "b"}}},
			want:     map[string]interface{}{"ssh_tunnel": map[string]interface{}{"ssh_host": "b"}},
		},
		{
			name:     "empty ssh tunnel and other lists are dropped",
			toFormat: "redash",
			in:       map[string]interface{}{"ssh_tunnel": []interface{}{}, "other": []interface{}{"x"}},
			want:     map[string]interface{}{},
		},
		{
			name:     "unknown format",
			toFormat: "yaml",
			in:       map[string]interface{}{"host": "h"},
			want:     map[string]interface{}{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertOptions(&tc.in, tc.toFormat); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDataSourceDataSource(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addDataSource(&fakeDataSource{
		Name:               "ACME Products",
		Type:               "pg",
		Paused:             1,
		PauseReason:        "maintenance",
		QueueName:          "queries",
		ScheduledQueueName: "scheduled_queries",
		Options:            map[string]interface{}{"host": "db.acme.com", "dbname": "products"},
	})

	d := schema.TestResourceDataRaw(t, dataSourceRedashDataSource().Schema, map[string]interface{}{"id": 1})
	if diags := dataSourceRedashDataSourceRead(ctx, d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}

	want := map[string]interface{}{
		"name":                 "ACME Products",
		"type":                 "pg",
		"syntax":               "sql",
		"paused":               1,
		"pause_reason":         "maintenance",
		"queue_name":           "queries",
		"scheduled_queue_name": "scheduled_queries",
		"options":              map[string]interface{}{"host": "db.acme.com", "dbname": "products"},
	}
	for k, v := range want {
		if got := d.Get(k); !reflect.DeepEqual(got, v) {
			t.Errorf("%s: got %v, want %v", k, got, v)
		}
	}
	if d.Id() != "1" {
		t.Errorf("expected id 1, got %q", d.Id())
	}

	d = schema.TestResourceDataRaw(t, dataSourceRedashDataSource().Schema, map[string]interface{}{"id": 99})
	diags := dataSourceRedashDataSourceRead(ctx, d, c)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "404") {
		t.Errorf("expected a 404 error for a missing data source, got %v", diags)
	}
}

func attachmentResourceData(t *testing.T, groupID, dataSourceID int) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceRedashGroupDataSourceAttachment().Schema, map[string]interface{}{
		"group_id":       groupID,
		"data_source_id": dataSourceID,
	})
}

func TestResourceGroupDataSourceAttachment_lifecycle(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	g := f.addGroup("Analysts")
	ds := f.addDataSource(&fakeDataSource{Name: "ACME Products", Type: "pg"})

	d := attachmentResourceData(t, g.ID, ds.ID)

	if diags := resourceRedashGroupDataSourceAttachmentCreate(ctx, d, c); diags.HasError() {
		t.Fatalf("create: %v", diags)
	}
	if !strings.HasPrefix(d.Id(), "1-1") {
		t.Errorf("expected id prefixed with 1-1, got %q", d.Id())
	}
	if _, ok := f.dataSource(ds.ID).Groups[g.ID]; !ok {
		t.Fatalf("expected data source to be attached to the group in Redash")
	}

	id := d.Id()
	if diags := resourceRedashGroupDataSourceAttachmentRead(ctx, d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != id {
		t.Errorf("expected an attached data source to stay in state, got id %q", d.Id())
	}

	if diags := resourceRedashGroupDataSourceAttachmentDelete(ctx, d, c); diags.HasError() {
		t.Fatalf("delete: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected id to be cleared, got %q", d.Id())
	}
	if _, ok := f.dataSource(ds.ID).Groups[g.ID]; ok {
		t.Errorf("expected data source to be detached from the group in Redash")
	}
}

func TestResourceGroupDataSourceAttachment_readRemovesDetached(t *testing.T) {
	f, c := newFakeRedash(t)
	g := f.addGroup("Analysts")
	ds := f.addDataSource(&fakeDataSource{Name: "ACME Products", Type: "pg"})

	d := attachmentResourceData(t, g.ID, ds.ID)
	d.SetId("1-1-detached")

	if diags := resourceRedashGroupDataSourceAttachmentRead(context.Background(), d, c); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected an attachment removed outside Terraform to be removed from state, got id %q", d.Id())
	}
}

func TestResourceGroupDataSourceAttachment_apiErrors(t *testing.T) {
	f, c := newFakeRedash(t)
	ctx := context.Background()
	f.addDataSource(&fakeDataSource{Name: "ACME Products", Type: "pg"})

	d := attachmentResourceData(t, 99, 1)
	diags := resourceRedashGroupDataSourceAttachmentCreate(ctx, d, c)
	if !diags.HasError() || !strings.Contains(diags[0].Summary, "404") {
		t.Errorf("expected create with a missing group to fail with 404, got %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("expected no id after a failed create, got %q", d.Id())
	}

	f.failWith = http.StatusInternalServerError
	d.SetId("99-1-x")
	for name, fn := range crudFuncs(resourceRedashGroupDataSourceAttachmentRead, nil, resourceRedashGroupDataSourceAttachmentDelete) {
		if diags := fn(ctx, d, c); !diags.HasError() {
			t.Errorf("expected %s to fail", name)
		}
	}
}
