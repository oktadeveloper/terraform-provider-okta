resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "web"
  grant_types    = ["implicit", "authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code", "token", "id_token"]
  issuer_mode    = "ORG_URL"

}

resource "okta_user" "test" {
  first_name = "TestAcc"
  last_name  = "Smith"
  login      = "testAcc_replace_with_uuid@example.com"
  email      = "testAcc_replace_with_uuid@example.com"
}

resource "okta_app_user" "test" {
  app_id            = okta_app_oauth.test.id
  user_id           = okta_user.test.id
  username          = okta_user.test.email
  retain_assignment = true
}
