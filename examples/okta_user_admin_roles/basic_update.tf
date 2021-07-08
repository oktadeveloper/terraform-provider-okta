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
    "USER_ADMIN",
    "HELP_DESK_ADMIN",
  ]
}
