package okta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAuthServerDefault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthServerDefaultUpdate,
		ReadContext:   resourceAuthServerDefaultRead,
		UpdateContext: resourceAuthServerDefaultUpdate,
		DeleteContext: resourceAuthServerDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"audiences": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Currently Okta only supports a single value here",
				Elem:        &schema.Schema{Type: schema.TypeString},
				DefaultFunc: func() (interface{}, error) {
					return []interface{}{"api://default"}, nil
				},
			},
			"status": statusSchema,
			"kid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials_last_rotated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials_next_rotation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials_rotation_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: elemInSlice([]string{"AUTO", "MANUAL"}),
				Description:      "Credential rotation mode, in many cases you cannot set this to MANUAL, the API will ignore the value and you will get a perpetual diff. This should rarely be used.",
				Default:          "MANUAL",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default Authorization Server for your Applications",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"issuer": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "allows you to use a custom issuer URL",
			},
			"issuer_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "*Early Access Property*. Indicates which value is specified in the issuer of the tokens that a Custom Authorization Server returns: the original Okta org domain URL or a custom domain URL",
				Default:          "ORG_URL",
				ValidateDiagFunc: elemInSlice([]string{"CUSTOM_URL", "ORG_URL"}),
			},
		},
	}
}

func resourceAuthServerDefaultRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	authServer, resp, err := getOktaClientFromMetadata(m).AuthorizationServer.GetAuthorizationServer(ctx, d.Id())
	if err := suppressErrorOn404(resp, err); err != nil {
		return diag.Errorf("failed to get authorization server: %v", err)
	}
	if authServer == nil {
		d.SetId("")
		return nil
	}
	_ = d.Set("audiences", convertStringSetToInterface(authServer.Audiences))
	if authServer.Credentials != nil && authServer.Credentials.Signing != nil {
		_ = d.Set("kid", authServer.Credentials.Signing.Kid)
		_ = d.Set("credentials_rotation_mode", authServer.Credentials.Signing.RotationMode)
		if authServer.Credentials.Signing.NextRotation != nil {
			_ = d.Set("credentials_next_rotation", authServer.Credentials.Signing.NextRotation.String())
		}
		if authServer.Credentials.Signing.LastRotated != nil {
			_ = d.Set("credentials_last_rotated", authServer.Credentials.Signing.LastRotated.String())
		}
	}
	_ = d.Set("description", authServer.Description)
	_ = d.Set("name", authServer.Name)
	_ = d.Set("status", authServer.Status)
	_ = d.Set("issuer", authServer.Issuer)

	// Do not sync these unless the issuer mode is specified since it is an EA feature and is computed in some cases
	if authServer.IssuerMode != "" {
		_ = d.Set("issuer_mode", authServer.IssuerMode)
	}
	return nil
}

func resourceAuthServerDefaultUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	if id == "" {
		id = d.Get("name").(string)
	}
	authServer, _, err := getOktaClientFromMetadata(m).AuthorizationServer.GetAuthorizationServer(ctx, id)
	if err != nil {
		return diag.Errorf("failed to get default authorization server: %v", err)
	}
	if status, ok := d.GetOk("status"); ok {
		client := getOktaClientFromMetadata(m)
		if status.(string) == statusActive && authServer.Status != statusActive {
			_, err := client.AuthorizationServer.ActivateAuthorizationServer(ctx, d.Id())
			if err != nil {
				return diag.Errorf("failed to activate default authorization server: %v", err)
			}
		}
		if status.(string) == statusInactive && authServer.Status != statusInactive {
			_, err := client.AuthorizationServer.DeactivateAuthorizationServer(ctx, d.Id())
			if err != nil {
				return diag.Errorf("failed to deactivate default authorization server: %v", err)
			}
		}
	}
	authServer.Audiences = convertInterfaceToStringSet(d.Get("audiences"))
	authServer.Credentials.Signing.RotationMode = d.Get("credentials_rotation_mode").(string)
	authServer.Description = d.Get("description").(string)
	authServer.Name = d.Get("name").(string)
	authServer.IssuerMode = d.Get("issuer_mode").(string)
	_, _, err = getOktaClientFromMetadata(m).AuthorizationServer.UpdateAuthorizationServer(ctx, id, *authServer)
	if err != nil {
		return diag.Errorf("failed to update default authorization server: %v", err)
	}
	d.SetId(authServer.Id)
	return resourceAuthServerDefaultRead(ctx, d, m)
}

// Default authorization server can not be removed
func resourceAuthServerDefaultDelete(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return nil
}
