# Data Source Data Source

Data source representation of an active Redash Data Source.

## Example Usage

```hcl
data "redash_data_source" "acme_corp" {
    id = 1
}

output "example" {
  value = "${jsonencode(data.redash_data_source.acme_corp)}"
}
```

## Argument Reference

* `id` - (Required) ID of Redash Data Source to load.

## Attribute Reference

* `id` - The ID of this Data Source
* `name` - The title of the Data Source
* `syntax` - The data syntax used for this Data Source
* `type` - The Data Source type
* `options` - An object storing the options for this Data Source
* `paused` - N/A
* `pause_reason` -  N/A
* `queue_name` -  N/A
* `scheduled_queue_name` -  N/A