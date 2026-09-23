# Alert Destination Resource

Allows creation/management of a Redash Alert Destination (where alert notifications are sent, e.g. email, Slack or a webhook).

## Example Usage

```hcl
resource "redash_alert_destination" "ops_email" {
  name = "Ops Email"
  type = "email"

  options = {
    addresses        = "ops@example.com"
    subject_template = "Redash alert: {alert_name}"
  }
}

resource "redash_alert_destination" "ops_webhook" {
  name = "Ops Webhook"
  type = "webhook"

  options = {
    url      = "https://hooks.example.com/redash"
    username = "redash"
    password = var.webhook_password
  }
}
```

## Argument Reference

* `name` - (Required) Name of the alert destination
* `type` - (Required) Destination type, e.g. `email`, `slack`, `webhook`, `mattermost`, `chatwork`, `pagerduty`, `hangouts_chat`
* `options` - (Required, Sensitive) Map of type-specific options. For the types listed below, only the listed options are accepted and any other option fails at plan time. Options for any other type are passed to Redash as-is.

| Type            | Options                                                  |
|-----------------|----------------------------------------------------------|
| `email`         | `addresses`, `subject_template`                          |
| `slack`         | `url`, `username`, `icon_emoji`, `icon_url`, `channel`   |
| `webhook`       | `url`, `username`, `password`                            |
| `mattermost`    | `url`, `username`, `icon_url`, `channel`                 |
| `chatwork`      | `api_token`, `room_id`, `message_template`               |
| `pagerduty`     | `integration_key`, `description`                         |
| `hangouts_chat` | `url`, `icon_url`                                        |

Redash never returns secret options (such as `password`, `api_token` or `integration_key`) once they are set, so the provider keeps the value from Terraform state. As a result, changes made to secrets outside Terraform are not detected.

## Attribute Reference

* `id` - Redash ID of this alert destination
* `icon` - Font Awesome icon Redash uses for this destination type

## Import

Alert destinations can be imported using their Redash ID:

```
terraform import redash_alert_destination.ops_email 1
```

Secret options can't be read back from Redash, so they are left out after import. Set them in your configuration and run `terraform apply` to bring state in line.
