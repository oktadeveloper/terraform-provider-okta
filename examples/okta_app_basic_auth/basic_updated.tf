resource "okta_user" "user" {
  admin_roles = [
    "APP_ADMIN",
  "USER_ADMIN"]
  first_name = "TestAcc"
  last_name  = "blah"
  login      = "testAcc-replace_with_uuid@example.com"
  email      = "testAcc-replace_with_uuid@example.com"
}

resource "okta_group" "group" {
  name = "testAcc_replace_with_uuid"
}

resource "okta_app_basic_auth" "test" {
  label    = "testAcc_replace_with_uuid"
  url      = "https://example.com/login.html"
  auth_url = "https://example.com/auth.html"
  logo     = "../examples/okta_app_basic_auth/terraform_icon.png"

  users {
    id       = okta_user.user.id
    username = okta_user.user.email
  }

  groups = [okta_group.group.id]
}
