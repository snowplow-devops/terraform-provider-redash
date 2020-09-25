# Data Source Resource

Allows creation/management of a Redash Data Source. As of writing, Redash supports 35 different source types, each with a different required options. The redash client library used by this provider uses your instance's API to handle available types and options, so please refer to documentation for your version. 

## Example Usage

```hcl
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

output "example" {
  value = "${jsonencode(data.redash_data_source.acme_corp)}"
}
```

## Argument Reference

* `name` - (Required) The title of the Data Source
* `type` - (Required) The Data Source type (check with Redash for latest)
* `options` - (Required) An object storing the options for this Data Source (check with Redash for required options by data source type)

## Attribute Reference

* `id` - The ID of this Data Source
* `name` - The title of the Data Source
* `syntax` - The data syntax used for this Data Source
* `type` - The Data Source type (Check with Redash for latest)
* `options` - An object storing the options for this Data Source
* `paused` - N/A
* `pause_reason` -  N/A
* `queue_name` -  N/A
* `scheduled_queue_name` -  N/A
