resource "okta_app_oauth" "test" {
  label          = "testAcc_replace_with_uuid"
  type           = "web"
  grant_types    = ["implicit", "authorization_code"]
  redirect_uris  = ["http://d.com/"]
  response_types = ["code", "token", "id_token"]
  issuer_mode    = "ORG_URL"

}

resource "okta_group" "test" {
  name = "testAcc_replace_with_uuid"
}

resource "okta_group" "test2" {
  name = "testAcc_replace_with_uuid_2"
}

resource "okta_group" "test3" {
  name = "testAcc_replace_with_uuid_3"
}

locals {
  group_ids = tolist([okta_group.test.id,okta_group.test2.id,okta_group.test3.id])
}

resource "okta_app_group_assignment" "test" {
  count = length(local.group_ids)

  app_id   = okta_app_oauth.test.id
  group_id = local.group_ids[count.index]
  priority = count.index
}
