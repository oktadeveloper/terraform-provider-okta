package okta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/terraform-provider-okta/sdk"
)

func resourceAuthServerPolicyDefault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthServerPolicyCreateDefault,
		ReadContext:   resourceAuthServerPolicyReadDefault,
		UpdateContext: resourceAuthServerPolicyUpdateDefault,
		DeleteContext: resourceAuthServerPolicyDeleteDefault,
		Importer:      createNestedResourceImporter([]string{"auth_server_id", "id"}),
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     sdk.OauthAuthorizationPolicyType,
				Description: "Auth server policy type, unlikely this will be anything other then the default",
				Deprecated:  "Policy type can only be of value 'OAUTH_AUTHORIZATION_POLICY', so this will be removed in the future, or set as 'Computed' value",
				DiffSuppressFunc: func(string, string, string, *schema.ResourceData) bool {
					return true
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auth_server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": statusSchema,
			"priority": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Priority of the auth server policy",
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_whitelist": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Use [\"ALL_CLIENTS\"] when unsure.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// resourceAuthServerPolicyCreateDefault imports by matching name, then applies the desired config via update.
func resourceAuthServerPolicyCreateDefault(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	respPolicy, err := findAuthPolicy(ctx, m, d.Get("auth_server_id").(string), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(respPolicy.Id)

	return resourceAuthServerPolicyUpdateDefault(ctx, d, m)
}

func resourceAuthServerPolicyReadDefault(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAuthServerPolicyRead(ctx, d, m)
}

func resourceAuthServerPolicyUpdateDefault(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAuthServerPolicyUpdate(ctx, d, m)
}

// resourceAuthServerPolicyDeleteDefault is no-op. Although Okta allows deleting the default policy,
// doing so here would break convention with other "default" type resources. Instead, no-op to avoid surprising the user.
func resourceAuthServerPolicyDeleteDefault(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func findAuthPolicy(ctx context.Context, m interface{}, serverID, name string) (*okta.Policy, error) {
	policies, _, err := getOktaClientFromMetadata(m).AuthorizationServer.ListAuthorizationServerPolicies(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("error listing policies for auth server '%s': %v", serverID, err)
	}
	for _, p := range policies {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("policy with name '%s' not found for auth server '%s'", name, serverID)
}
