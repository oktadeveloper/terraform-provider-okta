package okta

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

var baseAppSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Name of the app.",
	},
	"label": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Pretty name of app.",
	},
	"sign_on_mode": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Sign on mode of application.",
	},
	"status": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          statusActive,
		ValidateDiagFunc: stringInSlice([]string{statusActive, statusInactive}),
		Description:      "Status of application.",
	},
	"logo": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: logoValid(),
		Description:      "Logo of the application.",
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == ""
		},
	},
	"logo_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "URL of the application's logo",
	},
}

var appVisibilitySchema = map[string]*schema.Schema{
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
}

var baseAppSwaSchema = map[string]*schema.Schema{
	"accessibility_self_service": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable self service",
	},
	"accessibility_error_redirect_url": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Custom error page URL",
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
}

var appSamlDiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool {
	// Conditional default
	return new == "" && old == "http://www.okta.com/${org.externalKey}"
}

// Wish there was some better polymorphism that could make these similarities easier to deal with
func appRead(d *schema.ResourceData, name, status, signOn, label string, accy *okta.ApplicationAccessibility, vis *okta.ApplicationVisibility) {
	_ = d.Set("name", name)
	_ = d.Set("status", status)
	_ = d.Set("sign_on_mode", signOn)
	_ = d.Set("label", label)
	_ = d.Set("accessibility_self_service", *accy.SelfService)
	_ = d.Set("accessibility_error_redirect_url", accy.ErrorRedirectUrl)
	_ = d.Set("auto_submit_toolbar", vis.AutoSubmitToolbar)
	_ = d.Set("hide_ios", vis.Hide.IOS)
	_ = d.Set("hide_web", vis.Hide.Web)
}

func buildAppSchema(appSchema map[string]*schema.Schema) map[string]*schema.Schema {
	return buildSchema(baseAppSchema, appSchema)
}

func buildAppSchemaWithVisibility(appSchema map[string]*schema.Schema) map[string]*schema.Schema {
	return buildSchema(baseAppSchema, appVisibilitySchema, appSchema)
}

func buildSchemeCreds(d *schema.ResourceData) *okta.SchemeApplicationCredentials {
	revealPass := d.Get("reveal_password").(bool)

	return &okta.SchemeApplicationCredentials{
		RevealPassword: &revealPass,
		Scheme:         d.Get("credentials_scheme").(string),
		UserNameTemplate: &okta.ApplicationCredentialsUsernameTemplate{
			Template: d.Get("user_name_template").(string),
			Type:     d.Get("user_name_template_type").(string),
			Suffix:   d.Get("user_name_template_suffix").(string),
		},
		UserName: d.Get("shared_username").(string),
		Password: &okta.PasswordCredential{
			Value: d.Get("shared_password").(string),
		},
	}
}

func buildAppSwaSchema(appSchema map[string]*schema.Schema) map[string]*schema.Schema {
	return buildSchema(baseAppSchema, baseAppSwaSchema, appSchema)
}

func buildVisibility(d *schema.ResourceData) *okta.ApplicationVisibility {
	autoSubmit := d.Get("auto_submit_toolbar").(bool)
	hideMobile := d.Get("hide_ios").(bool)
	hideWeb := d.Get("hide_web").(bool)

	return &okta.ApplicationVisibility{
		AutoSubmitToolbar: &autoSubmit,
		Hide: &okta.ApplicationVisibilityHide{
			IOS: &hideMobile,
			Web: &hideWeb,
		},
	}
}

func fetchApp(ctx context.Context, d *schema.ResourceData, m interface{}, app okta.App) error {
	return fetchAppByID(ctx, d.Id(), m, app)
}

func fetchAppByID(ctx context.Context, id string, m interface{}, app okta.App) error {
	_, resp, err := getOktaClientFromMetadata(m).Application.GetApplication(ctx, id, app, nil)
	// We don't want to consider a 404 an error in some cases and thus the delineation.
	// Check if app's ID is set to ensure that app exists
	return suppressErrorOn404(resp, err)
}

func updateAppByID(ctx context.Context, id string, m interface{}, app okta.App) error {
	_, resp, err := getOktaClientFromMetadata(m).Application.UpdateApplication(ctx, id, app)
	// We don't want to consider a 404 an error in some cases and thus the delineation
	return suppressErrorOn404(resp, err)
}

func listApplicationGroupAssignments(ctx context.Context, client *okta.Client, id string) ([]*okta.ApplicationGroupAssignment, error) {
	var resGroups []*okta.ApplicationGroupAssignment
	groups, resp, err := client.Application.ListApplicationGroupAssignments(ctx, id, &query.Params{Limit: defaultPaginationLimit})
	if err != nil {
		return nil, err
	}
	for {
		resGroups = append(resGroups, groups...)
		if resp.HasNextPage() {
			resp, err = resp.Next(ctx, &groups)
			if err != nil {
				return nil, err
			}
			continue
		} else {
			break
		}
	}
	return resGroups, nil
}

func handleAppLogo(ctx context.Context, d *schema.ResourceData, m interface{}, appID string, links interface{}) error {
	l, ok := d.GetOk("logo")
	if !ok {
		return nil
	}
	_, err := getSupplementFromMetadata(m).UploadAppLogo(ctx, appID, l.(string))
	return err
}

func setAppStatus(ctx context.Context, d *schema.ResourceData, client *okta.Client, status string) error {
	desiredStatus := d.Get("status").(string)
	if status == desiredStatus {
		return nil
	}
	if desiredStatus == statusInactive {
		return responseErr(client.Application.DeactivateApplication(ctx, d.Id()))
	}
	return responseErr(client.Application.ActivateApplication(ctx, d.Id()))
}

// setAppSettings available preconfigured SAML and OAuth applications vary wildly on potential app settings, thus
// it is a generic map. This logic simply weeds out any empty string values.
func setAppSettings(d *schema.ResourceData, settings *okta.ApplicationSettingsApplication) error {
	flatMap := map[string]interface{}{}
	for key, val := range *settings {
		if str, ok := val.(string); ok {
			if str != "" {
				flatMap[key] = str
			}
		} else if val != nil {
			flatMap[key] = val
		}
	}
	payload, _ := json.Marshal(flatMap)
	return d.Set("app_settings_json", string(payload))
}

func setSamlSettings(d *schema.ResourceData, signOn *okta.SamlApplicationSettingsSignOn) error {
	_ = d.Set("default_relay_state", signOn.DefaultRelayState)
	_ = d.Set("sso_url", signOn.SsoAcsUrl)
	_ = d.Set("recipient", signOn.Recipient)
	_ = d.Set("destination", signOn.Destination)
	_ = d.Set("audience", signOn.Audience)
	_ = d.Set("idp_issuer", signOn.IdpIssuer)
	_ = d.Set("subject_name_id_template", signOn.SubjectNameIdTemplate)
	_ = d.Set("subject_name_id_format", signOn.SubjectNameIdFormat)
	_ = d.Set("response_signed", signOn.ResponseSigned)
	_ = d.Set("assertion_signed", signOn.AssertionSigned)
	_ = d.Set("signature_algorithm", signOn.SignatureAlgorithm)
	_ = d.Set("digest_algorithm", signOn.DigestAlgorithm)
	_ = d.Set("honor_force_authn", signOn.HonorForceAuthn)
	_ = d.Set("authn_context_class_ref", signOn.AuthnContextClassRef)
	if signOn.AllowMultipleAcsEndpoints != nil {
		if *signOn.AllowMultipleAcsEndpoints {
			acsEndpointsObj := signOn.AcsEndpoints
			if len(acsEndpointsObj) > 0 {
				acsEndpoints := make([]string, len(acsEndpointsObj))
				for i := range acsEndpointsObj {
					acsEndpoints[i] = acsEndpointsObj[i].Url
				}
				_ = d.Set("acs_endpoints", convertStringSetToInterface(acsEndpoints))
			}
		} else {
			_ = d.Set("acs_endpoints", nil)
		}
	}

	attrStatements := signOn.AttributeStatements
	arr := make([]map[string]interface{}, len(attrStatements))

	for i, st := range attrStatements {
		arr[i] = map[string]interface{}{
			"name":         st.Name,
			"namespace":    st.Namespace,
			"type":         st.Type,
			"values":       st.Values,
			"filter_type":  st.FilterType,
			"filter_value": st.FilterValue,
		}
	}
	if signOn.Slo != nil && signOn.Slo.Enabled != nil && *signOn.Slo.Enabled {
		_ = d.Set("single_logout_issuer", signOn.Slo.Issuer)
		_ = d.Set("single_logout_url", signOn.Slo.LogoutUrl)
		if len(signOn.SpCertificate.X5c) > 0 {
			_ = d.Set("single_logout_certificate", signOn.SpCertificate.X5c[0])
		}
	}
	return setNonPrimitives(d, map[string]interface{}{
		"attribute_statements": arr,
	})
}

func deleteApplication(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	client := getOktaClientFromMetadata(m)
	if d.Get("status").(string) == statusActive {
		_, err := client.Application.DeactivateApplication(ctx, d.Id())
		if err != nil {
			return err
		}
	}
	_, err := client.Application.DeleteApplication(ctx, d.Id())
	return err
}
