resource "okta_app_saml" "test" {
  label             = "testAcc_replace_with_uuid"
  preconfigured_app = "google"

  app_settings_json = <<JSON
  {
    "afwOnly": false,
    "domain": "articulate"
  }
JSON

}

resource "okta_user" "test" {
  first_name = "TestAcc"
  last_name  = "Smith"
  login      = "testAcc_replace_with_uuid@example.com"
  email      = "testAcc_replace_with_uuid@example.com"
}

resource "okta_app_user_schema" "test" {
  app_id      = okta_app_saml.test.id
  index       = "testCustom"
  title       = "terraform acceptance test"
  type        = "string"
  description = "terraform acceptance test updated"
  required    = true
  master      = "OKTA"
  scope       = "SELF"
}

resource "okta_app_user" "test" {
  app_id   = okta_app_saml.test.id
  user_id  = okta_user.test.id
  username = okta_user.test.email

  profile = <<JSON
{"testCustom":"testing"}
JSON

  depends_on = [okta_app_user_schema.test]
}
