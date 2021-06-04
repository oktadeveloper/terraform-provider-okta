package okta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func resourceAppBasicAuth() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppBasicAuthCreate,
		ReadContext:   resourceAppBasicAuthRead,
		UpdateContext: resourceAppBasicAuthUpdate,
		DeleteContext: resourceAppBasicAuthDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: buildAppSchemaWithVisibility(map[string]*schema.Schema{
			"auth_url": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Login button field",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Login password field",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
		}),
	}
}

func resourceAppBasicAuthCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	app := buildAppBasicAuth(d)
	activate := d.Get("status").(string) == statusActive
	params := &query.Params{Activate: &activate}
	_, _, err := client.Application.CreateApplication(ctx, app, params)
	if err != nil {
		return diag.Errorf("failed to create basic auth application: %v", err)
	}
	d.SetId(app.Id)
	err = handleAppLogo(ctx, d, m, app.Id, app.Links)
	if err != nil {
		return diag.Errorf("failed to upload logo for basic auth application: %v", err)
	}
	return resourceAppBasicAuthRead(ctx, d, m)
}

func resourceAppBasicAuthRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	app := okta.NewBasicAuthApplication()
	err := fetchApp(ctx, d, m, app)
	if err != nil {
		return diag.Errorf("failed to get basic auth application: %v", err)
	}
	if app.Id == "" {
		d.SetId("")
		return nil
	}
	_ = d.Set("url", app.Settings.App.Url)
	_ = d.Set("auth_url", app.Settings.App.AuthURL)
	_ = d.Set("name", app.Name)
	_ = d.Set("status", app.Status)
	_ = d.Set("sign_on_mode", app.SignOnMode)
	_ = d.Set("label", app.Label)
	_ = d.Set("auto_submit_toolbar", app.Visibility.AutoSubmitToolbar)
	_ = d.Set("hide_ios", app.Visibility.Hide.IOS)
	_ = d.Set("hide_web", app.Visibility.Hide.Web)
	_ = d.Set("logo_url", linksValue(app.Links, "logo", "href"))
	return nil
}

func resourceAppBasicAuthUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	app := buildAppBasicAuth(d)
	_, _, err := client.Application.UpdateApplication(ctx, d.Id(), app)
	if err != nil {
		return diag.Errorf("failed to update basic auth application: %v", err)
	}
	err = setAppStatus(ctx, d, client, app.Status)
	if err != nil {
		return diag.Errorf("failed to set basic auth application status: %v", err)
	}
	if d.HasChange("logo") {
		err = handleAppLogo(ctx, d, m, app.Id, app.Links)
		if err != nil {
			o, _ := d.GetChange("logo")
			_ = d.Set("logo", o)
			return diag.Errorf("failed to upload logo for basic auth application: %v", err)
		}
	}
	return resourceAppBasicAuthRead(ctx, d, m)
}

func resourceAppBasicAuthDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := deleteApplication(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to delete basic auth application: %v", err)
	}
	return nil
}

func buildAppBasicAuth(d *schema.ResourceData) *okta.BasicAuthApplication {
	// Abstracts away name and SignOnMode which are constant for this app type.
	app := okta.NewBasicAuthApplication()
	app.Label = d.Get("label").(string)

	app.Settings = &okta.BasicApplicationSettings{
		App: &okta.BasicApplicationSettingsApplication{
			AuthURL: d.Get("auth_url").(string),
			Url:     d.Get("url").(string),
		},
	}
	app.Visibility = buildVisibility(d)

	return app
}
