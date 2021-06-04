package okta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func resourceAppBookmark() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppBookmarkCreate,
		ReadContext:   resourceAppBookmarkRead,
		UpdateContext: resourceAppBookmarkUpdate,
		DeleteContext: resourceAppBookmarkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		// For those familiar with Terraform schemas be sure to check the base application schema and/or
		// the examples in the documentation
		Schema: buildAppSchemaWithVisibility(map[string]*schema.Schema{
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"request_integration": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		}),
	}
}

func resourceAppBookmarkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	app := buildAppBookmark(d)
	activate := d.Get("status").(string) == statusActive
	params := &query.Params{Activate: &activate}
	_, _, err := client.Application.CreateApplication(ctx, app, params)
	if err != nil {
		return diag.Errorf("failed to create bookmark application: %v", err)
	}
	d.SetId(app.Id)
	err = handleAppLogo(ctx, d, m, app.Id, app.Links)
	if err != nil {
		return diag.Errorf("failed to upload logo for bookmark application: %v", err)
	}
	return resourceAppBookmarkRead(ctx, d, m)
}

func resourceAppBookmarkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	app := okta.NewBookmarkApplication()
	err := fetchApp(ctx, d, m, app)
	if err != nil {
		return diag.Errorf("failed to get bookmark application: %v", err)
	}
	if app.Id == "" {
		d.SetId("")
		return nil
	}
	_ = d.Set("url", app.Settings.App.Url)
	_ = d.Set("label", app.Label)
	_ = d.Set("request_integration", app.Settings.App.RequestIntegration)
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

func resourceAppBookmarkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	app := buildAppBookmark(d)
	_, _, err := client.Application.UpdateApplication(ctx, d.Id(), app)
	if err != nil {
		return diag.Errorf("failed to update bookmark application: %v", err)
	}
	err = setAppStatus(ctx, d, client, app.Status)
	if err != nil {
		return diag.Errorf("failed to set bookmark application status: %v", err)
	}
	if d.HasChange("logo") {
		err = handleAppLogo(ctx, d, m, app.Id, app.Links)
		if err != nil {
			o, _ := d.GetChange("logo")
			_ = d.Set("logo", o)
			return diag.Errorf("failed to upload logo for bookmark application: %v", err)
		}
	}
	return resourceAppBookmarkRead(ctx, d, m)
}

func resourceAppBookmarkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := deleteApplication(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to delete bookmark application: %v", err)
	}
	return nil
}

func buildAppBookmark(d *schema.ResourceData) *okta.BookmarkApplication {
	app := okta.NewBookmarkApplication()
	integration := d.Get("request_integration").(bool)
	app.Label = d.Get("label").(string)
	app.Settings = &okta.BookmarkApplicationSettings{
		App: &okta.BookmarkApplicationSettingsApplication{
			RequestIntegration: &integration,
			Url:                d.Get("url").(string),
		},
	}
	app.Visibility = buildVisibility(d)
	return app
}
