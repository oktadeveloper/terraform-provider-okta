---
layout: "okta"
page_title: "Okta: okta_user_admin_roles"
sidebar_current: "docs-okta-resource-user-admin-roles"
description: |-
  Resource to manage a set of admin roles for a specific user.
---

# okta_user_admin_roles

Resource to manage a set of admin roles for a specific user.

This resource allows you to manage admin roles for a single user, independent of the user schema itself.

When using this with a `okta_user` resource, you should add a lifecycle ignore for admin roles to avoid conflicts
in desired state.

## Example Usage

```hcl
resource "okta_user" "test" {
  first_name = "TestAcc"
  last_name  = "Smith"
  login      = "testAcc-replace_with_uuid@example.com"
  email      = "testAcc-replace_with_uuid@example.com"

  lifecycle {
    ignore_changes = [admin_roles]
  }
}

resource "okta_user_admin_roles" "test" {
  user_id     = okta_user.test.id
  admin_roles = [
    "APP_ADMIN",
  ]
}
```

## Argument Reference

The following arguments are supported:

- `user_id` - (Required) ID of a Okta User.
- `admin_roles` - (Required) The list of Okta user admin roles, e.g. `["APP_ADMIN", "USER_ADMIN"]`

## Attributes Reference

N/A

## Import

Existing user admin roles can be imported via the Okta User ID.

```
$ terraform import okta_user_admin_roles.example <user id>
```
