# User Resource

Allows creation/management of a Redash User.

## Example Usage

```hcl
resource "redash_user" "wcoyote" {
  name   = "Wile E. Coyote"
  email  = "wcoyote@acme.com"
  groups = [3,2]
}

output "example" {
  value = "${jsonencode(data.redash_user.redash_user_rrunner)}"
}
```

## Argument Reference

* `name` - (Required) Full name of user
* `email` - (Required) Email address of user
* `groups` - (Optional) Array of group_ids user is a member of

## Attribute Reference

* `id` - User ID
* `name` - Full name of user
* `email` - Email address of user
* `auth_type` - Either "external" or "password" type
* `groups` - Array of group_ids user is a member of
* `profile_image_url` - Gravatar URL for user's profile image
* `is_invitation_pending` - Boolean if user has accepted invite yet
* `is_email_verified` - Boolean if user has verified email address yet
* `is_disabled` - Boolean if user has been disabled
* `active_at` - Timestamp of last activity
* `created_at` - Timestamp of create date
* `updated_at` - Timestamp of profile update
* `disabled_at` - Timestamp of when user was disabled