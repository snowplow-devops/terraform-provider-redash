# Redash Provider

Terraform provider for managing Redash configurations.

## Example Usage

```hcl
provider "redash" {
  redash_uri = "https://com.acme.redash"
}

resource "redash_data_source" "acme_corp" {
  name = "ACME Corporation Product Database"
  type = "redshift"

  options {
    host     = "newproducts.acme.com"
    port     = 5439
    dbname   = "products"
    user     = "wcoyote"
    password = "eth3LbeRt"
  }
}

resource "redash_group" "geniuses" {
  name = "ACME Users"
}

resource "redash_user" "wcoyote" {
  name   = "Wile E. Coyote"
  email  = "wcoyote@acme.com"
  groups = [redash_group.geniuses.id]

  depends_on = [
    redash_group.geniuses,
  ]
}

resource "redash_group_data_source_attachment" "wcoyote_acme" {
  group_id       = redash_group.geniuses.id
  data_source_id = redash_data_source.acme_corp.id

    depends_on = [
    redash_group.geniuses,
    redash_data_source.acme_corp,
  ]
}
```

## Argument Reference

* `redash_uri` (Required) The complete url to the instance you will be managing including protocol (e.g. https://acme.com/) 