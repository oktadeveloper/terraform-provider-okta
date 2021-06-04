resource "okta_auth_server_policy" "test" {
  status      = "ACTIVE"
  name        = "test"
  description = "test"
  priority    = 1
  client_whitelist = [
  "ALL_CLIENTS"]
  auth_server_id = okta_auth_server.test.id
}

resource "okta_auth_server" "test" {
  name        = "testAcc_replace_with_uuid"
  description = "test"
  audiences = [
  "whatever.rise.zone"]
}

data "okta_auth_server_policy" "test" {
  name           = "test"
  auth_server_id = okta_auth_server.test.id
}