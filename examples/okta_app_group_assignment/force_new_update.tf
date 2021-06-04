resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "web"
  grant_types    = ["implicit", "authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code", "token", "id_token"]
  issuer_mode    = "ORG_URL"

}

resource "okta_group" "test2" {
  name = "testAcc_replace_with_uuid-2"
}

resource "okta_app_group_assignment" "test" {
  app_id   = okta_app_oauth.test.id
  group_id = okta_group.test2.id
  priority = 1
}
