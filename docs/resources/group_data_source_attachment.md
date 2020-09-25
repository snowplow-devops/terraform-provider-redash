# Group Data Source Attachment Resource

The Group Data Source Attachment Resource allows management of Redash Data Sources that a Redash Group has permissions to.

## Example Usage

```hcl
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

* `group_id` - (Required) ID of Redash Group being modified
* `data_source_id` - (Required) ID of Redash Data Source to add to group
