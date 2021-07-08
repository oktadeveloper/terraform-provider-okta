---
layout: 'okta'
page_title: 'Okta: okta_user'
sidebar_current: 'docs-okta-resource-user'
description: |-
  Creates an Okta User.
---

# okta_user

Creates an Okta User.

This resource allows you to create and configure an Okta User.

## Example Usage

```hcl
resource "okta_user" "example" {
  first_name         = "John"
  last_name          = "Smith"
  login              = "john.smith@example.com"
  email              = "john.smith@example.com"
  city               = "New York"
  cost_center        = "10"
  country_code       = "US"
  department         = "IT"
  display_name       = "Dr. John Smith"
  division           = "Acquisitions"
  employee_number    = "111111"
  honorific_prefix   = "Dr."
  honorific_suffix   = "Jr."
  locale             = "en_US"
  manager            = "Jimbo"
  manager_id         = "222222"
  middle_name        = "John"
  mobile_phone       = "1112223333"
  nick_name          = "Johnny"
  organization       = "Testing Inc."
  postal_address     = "1234 Testing St."
  preferred_language = "en-us"
  primary_phone      = "4445556666"
  profile_url        = "http://www.example.com/profile"
  second_email       = "john.smith.fun@example.com"
  state              = "NY"
  street_address     = "5678 Testing Ave."
  timezone           = "America/New_York"
  title              = "Director"
  user_type          = "Employee"
  zip_code           = "11111"
}
```

## Argument Reference

The following arguments are supported:

- `email` - (Required) User profile property.

- `login` - (Required) User profile property.

- `first_name` - (Required) User's First Name, required by default.

- `last_name` - (Required) User's Last Name, required by default.

- `custom_profile_attributes` - (Optional) raw JSON containing all custom profile attributes.

- `admin_roles` - (Optional) Administrator roles assigned to User.
  - `DEPRECATED`: Please replace usage with the `okta_user_admin_roles` resource.

- `city` - (Optional) User profile property.

- `cost_center` - (Optional) User profile property.

- `country_code` - (Optional) User profile property.

- `department` - (Optional) User profile property.

- `display_name` - (Optional) User profile property.

- `division` - (Optional) User profile property.

- `employee_number` - (Optional) User profile property.

- `group_memberships` - (Optional) User profile property.

- `honorific_prefix` - (Optional) User profile property.

- `honorific_suffix` - (Optional) User profile property.

- `locale` - (Optional) User profile property.

- `manager` - (Optional) User profile property.

- `manager_id` - (Optional) User profile property.

- `middle_name` - (Optional) User profile property.

- `mobile_phone` - (Optional) User profile property.

- `nick_name` - (Optional) User profile property.

- `organization` - (Optional) User profile property.

- `postal_address` - (Optional) User profile property.

- `preferred_language` - (Optional) User profile property.

- `primary_phone` - (Optional) User profile property.

- `profile_url` - (Optional) User profile property.

- `second_email` - (Optional) User profile property.

- `state` - (Optional) User profile property.

- `status` - (Optional) User profile property.

- `street_address` - (Optional) User profile property.

- `timezone` - (Optional) User profile property.

- `title` - (Optional) User profile property.

- `user_type` - (Optional) User profile property.

- `zip_code` - (Optional) User profile property.

- `password` - (Optional) User password.

- `recovery_question` - (Optional) User password recovery question.

- `recovery_answer` - (Optional) User password recovery answer.

## Attributes Reference

- `id` - (Optional) ID of the User schema property.

## Import

An Okta User can be imported via the ID.

```
$ terraform import okta_user.example <user id>
```
