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

func dataSourceAppOauth() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppOauthRead,
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
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_submit_toolbar": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Display auto submit toolbar",
			},
			"hide_ios": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Do not display application icon on mobile app",
			},
			"hide_web": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Do not display application icon to users",
			},
			"grant_types": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "List of OAuth 2.0 grant types",
			},
			"response_types": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "List of OAuth 2.0 response type strings.",
			},
			"redirect_uris": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "List of URIs for use in the redirect-based flow.",
			},
			"post_logout_redirect_uris": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "List of URIs for redirection after logout",
			},
			"logo_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI that references a logo for the client.",
			},
			"login_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI that initiates login.",
			},
			"login_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of Idp-Initiated login that the client supports, if any",
			},
			"login_scopes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of scopes to use for the request when 'login_mode' == OKTA",
			},
			"client_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI to a web page providing information about the client.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "OAuth client ID",
			},
			"policy_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URI to web page providing client policy document.",
			},
			"links": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Discoverable resources related to the app",
			},
		},
	}
}

func dataSourceAppOauthRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	filters, err := getAppFilters(d)
	if err != nil {
		return diag.Errorf("invalid OAuth app filters: %v", err)
	}
	var app *okta.OpenIdConnectApplication
	if filters.ID != "" {
		respApp, _, err := getOktaClientFromMetadata(m).Application.GetApplication(ctx, filters.ID, okta.NewOpenIdConnectApplication(), nil)
		if err != nil {
			return diag.Errorf("failed get app by ID: %v", err)
		}
		app = respApp.(*okta.OpenIdConnectApplication)
	} else {
		re := getOktaClientFromMetadata(m).GetRequestExecutor()
		qp := &query.Params{Limit: 1, Filter: filters.Status, Q: filters.getQ()}
		req, err := re.NewRequest("GET", fmt.Sprintf("/api/v1/apps%s", qp.String()), nil)
		if err != nil {
			return diag.Errorf("failed to list OAuth apps: %v", err)
		}
		var appList []*okta.OpenIdConnectApplication
		_, err = re.Do(ctx, req, &appList)
		if err != nil {
			return diag.Errorf("failed to list OAuth apps: %v", err)
		}
		if len(appList) < 1 {
			return diag.Errorf("no OAuth application found with provided filter: %s", filters)
		}
		if filters.Label != "" && appList[0].Label != filters.Label {
			return diag.Errorf("no OAuth application found with the provided label: %s", filters.Label)
		}
		logger(m).Info("found multiple OAuth applications with the criteria supplied, using the first one, sorted by creation date")
		app = appList[0]
	}
	_ = d.Set("label", app.Label)
	_ = d.Set("name", app.Name)
	_ = d.Set("status", app.Status)
	_ = d.Set("type", app.Settings.OauthClient.ApplicationType)
	_ = d.Set("auto_submit_toolbar", app.Visibility.AutoSubmitToolbar)
	_ = d.Set("hide_ios", app.Visibility.Hide.IOS)
	_ = d.Set("hide_web", app.Visibility.Hide.Web)
	_ = d.Set("client_uri", app.Settings.OauthClient.ClientUri)
	_ = d.Set("logo_uri", app.Settings.OauthClient.LogoUri)
	_ = d.Set("login_uri", app.Settings.OauthClient.InitiateLoginUri)
	_ = d.Set("client_id", app.Credentials.OauthClient.ClientId)
	_ = d.Set("policy_uri", app.Settings.OauthClient.PolicyUri)
	respTypes := make([]string, len(app.Settings.OauthClient.ResponseTypes))
	for i := range app.Settings.OauthClient.ResponseTypes {
		respTypes[i] = string(*app.Settings.OauthClient.ResponseTypes[i])
	}
	grantTypes := make([]string, len(app.Settings.OauthClient.GrantTypes))
	for i := range app.Settings.OauthClient.GrantTypes {
		grantTypes[i] = string(*app.Settings.OauthClient.GrantTypes[i])
	}
	aggMap := map[string]interface{}{
		"redirect_uris":             convertStringSetToInterface(app.Settings.OauthClient.RedirectUris),
		"response_types":            convertStringSetToInterface(respTypes),
		"grant_types":               convertStringSetToInterface(grantTypes),
		"post_logout_redirect_uris": convertStringSetToInterface(app.Settings.OauthClient.PostLogoutRedirectUris),
	}
	if app.Settings.OauthClient.IdpInitiatedLogin != nil {
		_ = d.Set("login_mode", app.Settings.OauthClient.IdpInitiatedLogin.Mode)
		aggMap["login_scopes"] = convertStringSetToInterface(app.Settings.OauthClient.IdpInitiatedLogin.DefaultScope)
	}
	err = setNonPrimitives(d, aggMap)
	if err != nil {
		return diag.Errorf("failed to set OAuth application properties: %v", err)
	}
	p, _ := json.Marshal(app.Links)
	_ = d.Set("links", string(p))
	return nil
}
