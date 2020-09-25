# Group Data Source

Data source representation of an active Redash Group

## Example Usage

```hcl
data "redash_group" "geniuses" {
    id = 1
}

output "example" {
  value = "${jsonencode(data.redash_group.geniuses)}"
}
```

## Argument Reference

* `id` - (Required) Group ID to load

## Attribute Reference

* `id` - Redash ID of this group
* `name` - Redash ID of this group
* `type` - "builtin" or "regular" - built-in groups cannot be modified
* `permissions` - CSV of available permissions to group
* `created_at` - Timestamp of group creation