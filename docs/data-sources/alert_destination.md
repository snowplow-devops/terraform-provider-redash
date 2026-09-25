# Alert Destination Data Source

Data source representation of an existing Redash Alert Destination

## Example Usage

```hcl
data "redash_alert_destination" "ops_email" {
  name = "Ops Email"
}

output "example" {
  value = "${jsonencode(data.redash_alert_destination.ops_email)}"
}
```

## Argument Reference

Exactly one of the following must be set:

* `id` - Alert destination ID to load
* `name` - Alert destination name to load. Fails if more than one destination has this name.

## Attribute Reference

* `id` - Redash ID of this alert destination
* `name` - Name of this alert destination
* `type` - Destination type, e.g. `email`, `slack`, `webhook`
* `icon` - Font Awesome icon Redash uses for this destination type

Options are not exposed, as Redash does not return secret values.
