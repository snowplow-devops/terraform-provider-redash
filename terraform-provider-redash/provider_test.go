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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/snowplow-devops/redash-client-go/redash"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"redash": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func TestProviderConfigure(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"api_key":    "test-key",
		"redash_uri": "https://redash.example.com",
	})

	meta, diags := providerConfigure(context.Background(), d)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	c, ok := meta.(*redash.Client)
	if !ok || c.Config.APIKey != "test-key" || c.Config.RedashURI != "https://redash.example.com" {
		t.Errorf("unexpected client: %#v", meta)
	}
}

func TestProviderConfigure_invalidURI(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"api_key":    "test-key",
		"redash_uri": "ftp://redash.example.com",
	})

	if _, diags := providerConfigure(context.Background(), d); !diags.HasError() {
		t.Errorf("expected an error for a non-HTTP URI")
	}
}
