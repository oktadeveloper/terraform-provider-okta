package okta

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func dataSourceAppSaml() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppSamlRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"label", "label_prefix"},
			},
			"label": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "label_prefix"},
			},
			"label_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "label"},
			},
			"active_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Search only ACTIVE applications.",
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
				Type:        schema.TypeString,
				Description: "Certificate ID",
				Computed:    true,
			},
			"auto_submit_toolbar": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Display auto submit toolbar",
			},
			"hide_ios": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Do not display application icon on mobile app",
			},
			"hide_web": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Do not display application icon to users",
			},
			"default_relay_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifies a specific application resource in an IDP initiated SSO scenario.",
			},
			"sso_url": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Single Sign On URL",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"recipient": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The location where the app may present the SAML assertion",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"destination": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Identifies the location where the SAML response is intended to be sent inside of the SAML assertion",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"audience": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Audience Restriction",
			},
			"idp_issuer": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "SAML issuer ID",
				DiffSuppressFunc: appSamlDiffSuppressFunc,
			},
			"sp_issuer": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SAML SP issuer ID",
			},
			"subject_name_id_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Template for app user's username when a user is assigned to the app",
			},
			"subject_name_id_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifies the SAML processing rules.",
				ValidateDiagFunc: stringInSlice(
					[]string{
						"urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
						"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
						"urn:oasis:names:tc:SAML:1.1:nameid-format:x509SubjectName",
						"urn:oasis:names:tc:SAML:2.0:nameid-format:persistent",
						"urn:oasis:names:tc:SAML:2.0:nameid-format:transient",
					},
				),
			},
			"response_signed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Determines whether the SAML auth response message is digitally signed",
			},
			"request_compressed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Denotes whether the request is compressed or not.",
			},
			"assertion_signed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Determines whether the SAML assertion is digitally signed",
			},
			"signature_algorithm": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Signature algorithm used ot digitally sign the assertion and response",
				ValidateDiagFunc: stringInSlice([]string{"RSA_SHA256", "RSA_SHA1"}),
			},
			"digest_algorithm": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Determines the digest algorithm used to digitally sign the SAML assertion and response",
				ValidateDiagFunc: stringInSlice([]string{"SHA256", "SHA1"}),
			},
			"honor_force_authn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Prompt user to re-authenticate if SP asks for it",
			},
			"authn_context_class_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifies the SAML authentication context class for the assertion’s authentication statement",
			},
			"accessibility_self_service": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable self service",
			},
			"accessibility_error_redirect_url": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Custom error page URL",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"accessibility_login_redirect_url": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Custom login page URL",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"features": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "features to enable",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"user_name_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "${source.login}",
				Description: "Username template",
			},
			"user_name_template_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username template suffix",
			},
			"user_name_template_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "BUILT_IN",
				Description:      "Username template type",
				ValidateDiagFunc: stringInSlice([]string{"NONE", "CUSTOM", "BUILT_IN"}),
			},
			"app_settings_json": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Application settings in JSON format",
				ValidateDiagFunc: stringIsJSON,
				StateFunc:        normalizeDataJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == ""
				},
			},
			"acs_endpoints": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of ACS endpoints for this SAML application",
			},
			"attribute_statements": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of group attribute filter",
						},
						"filter_value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Filter value to use",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reference name of the attribute statement",
						},
						"namespace": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name format of the attribute",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of attribute statements object",
						},
						"values": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"single_logout_issuer": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The issuer of the Service Provider that generates the Single Logout request",
			},
			"single_logout_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The location where the logout response is sent",
			},
			"single_logout_certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "x509 encoded certificate that the Service Provider uses to sign Single Logout requests",
			},
			"links": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Discoverable resources related to the app",
			},
		},
	}
}

func dataSourceAppSamlRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	filters, err := getAppFilters(d)
	if err != nil {
		return diag.Errorf("invalid SAML app filters: %v", err)
	}
	var app *okta.SamlApplication
	if filters.ID != "" {
		respApp, _, err := getOktaClientFromMetadata(m).Application.GetApplication(ctx, filters.ID, okta.NewSamlApplication(), nil)
		if err != nil {
			return diag.Errorf("failed get app by ID: %v", err)
		}
		app = respApp.(*okta.SamlApplication)
	} else {
		re := getOktaClientFromMetadata(m).GetRequestExecutor()
		qp := &query.Params{Limit: 1, Filter: filters.Status, Q: filters.getQ()}
		req, err := re.NewRequest("GET", fmt.Sprintf("/api/v1/apps%s", qp.String()), nil)
		if err != nil {
			return diag.Errorf("failed to list SAML apps: %v", err)
		}
		var appList []*okta.SamlApplication
		_, err = re.Do(ctx, req, &appList)
		if err != nil {
			return diag.Errorf("failed to list SAML apps: %v", err)
		}
		if len(appList) < 1 {
			return diag.Errorf("no SAML application found with provided filter: %s", filters)
		}
		if filters.Label != "" && appList[0].Label != filters.Label {
			return diag.Errorf("no SAML application found with the provided label: %s", filters.Label)
		}
		logger(m).Info("found multiple SAML applications with the criteria supplied, using the first one, sorted by creation date")
		app = appList[0]
	}

	d.SetId(app.Id)
	_ = d.Set("label", app.Label)
	_ = d.Set("name", app.Name)
	_ = d.Set("status", app.Status)
	_ = d.Set("key_id", app.Credentials.Signing.Kid)
	if app.Settings != nil {
		if app.Settings.SignOn != nil {
			err = setSamlSettings(d, app.Settings.SignOn)
			if err != nil {
				return diag.Errorf("failed to read SAML app: error setting SAML sign-on settings: %v", err)
			}
		}
		err = setAppSettings(d, app.Settings.App)
		if err != nil {
			return diag.Errorf("failed to read SAML app: failed to set SAML app settings: %v", err)
		}
	}
	_ = d.Set("features", convertStringSetToInterface(app.Features))
	_ = d.Set("user_name_template", app.Credentials.UserNameTemplate.Template)
	_ = d.Set("user_name_template_type", app.Credentials.UserNameTemplate.Type)
	_ = d.Set("user_name_template_suffix", app.Credentials.UserNameTemplate.Suffix)
	p, _ := json.Marshal(app.Links)
	_ = d.Set("links", string(p))
	return nil
}
