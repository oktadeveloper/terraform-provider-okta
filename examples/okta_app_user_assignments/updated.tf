resource "okta_app_oauth" "test" {
  label                      = "testAcc_replace_with_uuid"
  type                       = "browser"
  grant_types                = ["authorization_code"]
  token_endpoint_auth_method = "none"
  redirect_uris              = ["https://testing.com"]
  response_types             = ["code"]

  lifecycle {
    ignore_changes = ["users", "groups"]
  }
}

resource "okta_user" "test1" {
  first_name = "Test"
  last_name  = "Broker"
  login      = "testAcc_broker_replace_with_uuid@example.com"
  email      = "testAcc_broker_replace_with_uuid@example.com"
}

resource "okta_user" "test2" {
  first_name = "Test"
  last_name  = "Python"
  login      = "testAcc_python_replace_with_uuid@example.com"
  email      = "testAcc_python_replace_with_uuid@example.com"
}

resource "okta_user" "test3" {
  first_name = "Test"
  last_name  = "Jaeger"
  login      = "testAcc_jaeger_replace_with_uuid@example.com"
  email      = "testAcc_jaeger_replace_with_uuid@example.com"
}

resource "okta_app_user_assignments" "test" {
  app_id = okta_app_oauth.test.id

  users {
    id       = okta_user.test1.id
    username = okta_user.test1.login
  }

  users {
    id       = okta_user.test2.id
    username = okta_user.test2.login
  }
}
