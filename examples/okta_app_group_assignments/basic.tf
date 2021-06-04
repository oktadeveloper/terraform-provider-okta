resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "web"
  grant_types    = ["implicit", "authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code", "token", "id_token"]
  issuer_mode    = "ORG_URL"

  lifecycle {
    ignore_changes = ["users", "groups"]
  }
}

resource "okta_group" "test1" {
  name = "testAcc_replace_with_uuid"
}

resource "okta_group" "test2" {
  name = "testAcc_replace_with_uuid_2"
}

resource "okta_group" "test3" {
  name = "testAcc_replace_with_uuid_3"
}

resource "okta_app_group_assignments" "test" {
  app_id = okta_app_oauth.test.id

  group {
    id       = okta_group.test1.id
    priority = 1
  }
  group {
    id       = okta_group.test2.id
    priority = 2
  }
  group {
    id       = okta_group.test3.id
    priority = 3
  }
}
