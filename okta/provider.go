// Package okta terraform configuration for an okta site
package okta

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource names, defined in place, used throughout the provider and tests
const (
	appAutoLogin           = "okta_app_auto_login"
	appBookmark            = "okta_app_bookmark"
	appBasicAuth           = "okta_app_basic_auth"
	appGroupAssignment     = "okta_app_group_assignment"
	appUser                = "okta_app_user"
	appOAuth               = "okta_app_oauth"
	appOAuthRedirectURI    = "okta_app_oauth_redirect_uri"
	appSaml                = "okta_app_saml"
	appSecurePasswordStore = "okta_app_secure_password_store"
	appSwa                 = "okta_app_swa"
	appThreeField          = "okta_app_three_field"
	appUserSchema          = "okta_app_user_schema"
	appUserBaseSchema      = "okta_app_user_base_schema"
	authServer             = "okta_auth_server"
	authServerClaim        = "okta_auth_server_claim"
	authServerPolicy       = "okta_auth_server_policy"
	authServerPolicyRule   = "okta_auth_server_policy_rule"
	authServerScope        = "okta_auth_server_scope"
	eventHook              = "okta_event_hook"
	factor                 = "okta_factor"
	groupRoles             = "okta_group_roles"
	groupRule              = "okta_group_rule"
	idpOidc                = "okta_idp_oidc"
	idpSaml                = "okta_idp_saml"
	idpSamlKey             = "okta_idp_saml_key"
	idpSocial              = "okta_idp_social"
	inlineHook             = "okta_inline_hook"
	networkZone            = "okta_network_zone"
	oktaGroup              = "okta_group"
	oktaProfileMapping     = "okta_profile_mapping"
	oktaUser               = "okta_user"
	policyMfa              = "okta_policy_mfa"
	policyPassword         = "okta_policy_password"
	policyRuleIdpDiscovery = "okta_policy_rule_idp_discovery"
	policyRuleMfa          = "okta_policy_rule_mfa"
	policyRulePassword     = "okta_policy_rule_password"
	policyRuleSignOn       = "okta_policy_rule_signon"
	policySignOn           = "okta_policy_signon"
	templateEmail          = "okta_template_email"
	templateSms            = "okta_template_sms"
	trustedOrigin          = "okta_trusted_origin"
	userBaseSchema         = "okta_user_base_schema"
	userSchema             = "okta_user_schema"
	userType               = "okta_user_type"
)

// Provider establishes a client connection to an okta site
// determined by its schema string values
func Provider() *schema.Provider {
	deprecatedPolicies := dataSourceDefaultPolicies()
	deprecatedPolicies.DeprecationMessage = "This data source will be deprecated in favor of okta_default_policy or okta_policy data sources."

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"org_name": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_ORG_NAME", nil),
				Description: "The organization to manage in Okta.",
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_API_TOKEN", nil),
				Description: "API Token granting privileges to Okta API.",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OKTA_BASE_URL", "okta.com"),
				Description: "The Okta url. (Use 'oktapreview.com' for Okta testing).",
			},
			"backoff": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Use exponential back off strategy for rate limits.",
			},
			"min_wait_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Minimum seconds to wait when the rate limit is hit. We use exponential backoffs when backoff is enabled.",
			},
			"max_wait_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "Maximum seconds to wait when the rate limit is hit. We use exponential backoffs when backoff is enabled.",
			},
			"max_retries": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          5,
				ValidateDiagFunc: intAtMost(100), // Have to cut it off somewhere right?
				Description:      "Maximum number of retries to attempt before erroring out.",
			},
			"parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Number of concurrent requests to make within a resource where bulk operations are not possible. Take note of https://developer.okta.com/docs/api/getting_started/rate-limits.",
			},
			"log_level": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          int(hclog.Error),
				ValidateDiagFunc: intBetween(1, 5),
				Description:      "providers log level. Minimum is 1 (TRACE), and maximum is 5 (ERROR)",
			},
			"max_requests": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          100,
				ValidateDiagFunc: intBetween(1, 100),
				Description: "(Experimental) controls how many requests can be made to each Okta endpoint by the provider. " +
					"It is used to prevent rate limit violations. By default request throttling is disabled meaning the provider " +
					"might cause rate limits violations. Expects an integer representing a percentage value - e.g. `40`. " +
					"`40` means that the provider is allowed to use up to 40% of the rate limit. E.g. assuming rate limit for " +
					"`/api/v1/apps` endpoint is 25, up to 10 requests will be made that burn `/api/v1/apps` rate limit. " +
					"Currently request throttling works only for `/api/v1/apps` rate limit.",
			},
			"request_timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          0,
				ValidateDiagFunc: intBetween(0, 100),
				Description:      "Timeout for single request (in seconds) which is made to Okta, the default is `0` (means no limit is set). The maximum value can be `100`.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			appAutoLogin:           resourceAppAutoLogin(),
			appBookmark:            resourceAppBookmark(),
			appBasicAuth:           resourceAppBasicAuth(),
			appGroupAssignment:     resourceAppGroupAssignment(),
			appUser:                resourceAppUser(),
			appOAuth:               resourceAppOAuth(),
			appOAuthRedirectURI:    resourceAppOAuthRedirectURI(),
			appSaml:                resourceAppSaml(),
			appSecurePasswordStore: resourceAppSecurePasswordStore(),
			appSwa:                 resourceAppSwa(),
			appThreeField:          resourceAppThreeField(),
			appUserSchema:          resourceAppUserSchema(),
			appUserBaseSchema:      resourceAppUserBaseSchema(),
			authServer:             resourceAuthServer(),
			authServerClaim:        resourceAuthServerClaim(),
			authServerPolicy:       resourceAuthServerPolicy(),
			authServerPolicyRule:   resourceAuthServerPolicyRule(),
			authServerScope:        resourceAuthServerScope(),
			eventHook:              resourceEventHook(),
			factor:                 resourceFactor(),
			groupRoles:             resourceGroupRoles(),
			groupRule:              resourceGroupRule(),
			idpOidc:                resourceIdpOidc(),
			idpSaml:                resourceIdpSaml(),
			idpSamlKey:             resourceIdpSigningKey(),
			idpSocial:              resourceIdpSocial(),
			inlineHook:             resourceInlineHook(),
			networkZone:            resourceNetworkZone(),
			oktaGroup:              resourceGroup(),
			oktaProfileMapping:     resourceOktaProfileMapping(),
			oktaUser:               resourceUser(),
			policyMfa:              resourcePolicyMfa(),
			policyPassword:         resourcePolicyPassword(),
			policySignOn:           resourcePolicySignOn(),
			policyRuleIdpDiscovery: resourcePolicyRuleIdpDiscovery(),
			policyRuleMfa:          resourcePolicyMfaRule(),
			policyRulePassword:     resourcePolicyPasswordRule(),
			policyRuleSignOn:       resourcePolicySignonRule(),
			templateEmail:          resourceTemplateEmail(),
			templateSms:            resourceTemplateSms(),
			trustedOrigin:          resourceTrustedOrigin(),
			userSchema:             resourceUserSchema(),
			userBaseSchema:         resourceUserBaseSchema(),
			userType:               resourceUserType(),

			// The day I realized I was naming stuff wrong :'-(
			"okta_idp":                       deprecateIncorrectNaming(resourceIdpOidc(), idpOidc),
			"okta_saml_idp":                  deprecateIncorrectNaming(resourceIdpSaml(), idpSaml),
			"okta_saml_idp_signing_key":      deprecateIncorrectNaming(resourceIdpSigningKey(), idpSamlKey),
			"okta_social_idp":                deprecateIncorrectNaming(resourceIdpSocial(), idpSocial),
			"okta_bookmark_app":              deprecateIncorrectNaming(resourceAppBookmark(), appBookmark),
			"okta_saml_app":                  deprecateIncorrectNaming(resourceAppSaml(), appSaml),
			"okta_oauth_app":                 deprecateIncorrectNaming(resourceAppOAuth(), appOAuth),
			"okta_oauth_app_redirect_uri":    deprecateIncorrectNaming(resourceAppOAuthRedirectURI(), appOAuthRedirectURI),
			"okta_auto_login_app":            deprecateIncorrectNaming(resourceAppAutoLogin(), appAutoLogin),
			"okta_secure_password_store_app": deprecateIncorrectNaming(resourceAppSecurePasswordStore(), appSecurePasswordStore),
			"okta_three_field_app":           deprecateIncorrectNaming(resourceAppThreeField(), appThreeField),
			"okta_swa_app":                   deprecateIncorrectNaming(resourceAppSwa(), appSwa),
			"okta_password_policy":           deprecateIncorrectNaming(resourcePolicyPassword(), policyPassword),
			"okta_signon_policy":             deprecateIncorrectNaming(resourcePolicySignOn(), policySignOn),
			"okta_signon_policy_rule":        deprecateIncorrectNaming(resourcePolicySignonRule(), policyRuleSignOn),
			"okta_password_policy_rule":      deprecateIncorrectNaming(resourcePolicyPasswordRule(), policyRulePassword),
			"okta_mfa_policy":                deprecateIncorrectNaming(resourcePolicyMfa(), policyMfa),
			"okta_mfa_policy_rule":           deprecateIncorrectNaming(resourcePolicyMfaRule(), policyRuleMfa),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"okta_app":                         dataSourceApp(),
			"okta_app_saml":                    dataSourceAppSaml(),
			"okta_app_metadata_saml":           dataSourceAppMetadataSaml(),
			"okta_default_policies":            deprecatedPolicies,
			"okta_default_policy":              dataSourceDefaultPolicies(),
			"okta_everyone_group":              dataSourceEveryoneGroup(),
			"okta_group":                       dataSourceGroup(),
			"okta_idp_metadata_saml":           dataSourceIdpMetadataSaml(),
			"okta_idp_saml":                    dataSourceIdpSaml(),
			idpOidc:                            dataSourceIdpOidc(),
			"okta_policy":                      dataSourcePolicy(),
			"okta_user_profile_mapping_source": dataSourceUserProfileMappingSource(),
			"okta_user":                        dataSourceUser(),
			"okta_users":                       dataSourceUsers(),
			authServer:                         dataSourceAuthServer(),
			userType:                           dataSourceUserType(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func deprecateIncorrectNaming(d *schema.Resource, newResource string) *schema.Resource {
	d.DeprecationMessage = fmt.Sprintf("Resource is deprecated due to a correction in naming conventions, please use %s instead.", newResource)
	return d
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Printf("[INFO] Initializing Okta client")

	config := Config{
		orgName:        d.Get("org_name").(string),
		domain:         d.Get("base_url").(string),
		apiToken:       d.Get("api_token").(string),
		parallelism:    d.Get("parallelism").(int),
		retryCount:     d.Get("max_retries").(int),
		maxRequests:    d.Get("max_requests").(int),
		minWait:        d.Get("min_wait_seconds").(int),
		maxWait:        d.Get("max_wait_seconds").(int),
		backoff:        d.Get("backoff").(bool),
		logLevel:       d.Get("log_level").(int),
		requestTimeout: d.Get("request_timeout").(int),
	}
	if err := config.loadAndValidate(); err != nil {
		return nil, diag.Errorf("[ERROR] Error initializing the Okta SDK clients: %v", err)
	}
	return &config, nil
}
