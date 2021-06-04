resource "okta_policy_mfa_default" "test" {
  google_otp = {
    enroll = "OPTIONAL"
  }
  depends_on = [okta_factor.google_otp]
}

resource "okta_factor" "google_otp" {
  provider_id = "google_otp"
}
