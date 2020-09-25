# Group Resource

Allows creation/management of a Redash Group.

## Example Usage

```hcl
resource "redash_group" "runners" {
  name = "Beep Beep"
}

output "example" {
  value = "${jsonencode(data.redash_group.runners)}"
}
```

## Argument Reference

* `name` - (Required) List arguments this resource takes.

## Attribute Reference

* `id` - Redash ID of this group
* `name` - Redash ID of this group
* `type` - "builtin" or "regular" - built-in groups cannot be modified
* `permissions` - CSV of available permissions to group
* `created_at` - Timestamp of group creation