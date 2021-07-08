package okta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserAdminRoles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAdminRolesCreate,
		ReadContext:   resourceUserAdminRolesRead,
		UpdateContext: resourceUserAdminRolesUpdate,
		DeleteContext: resourceUserAdminRolesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("user_id", d.Id())
				d.SetId(d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Description: "Resource to manage a set of administrator roles for a specific user.",
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of a Okta User",
				ForceNew:    true,
			},
			"admin_roles": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "User Okta admin roles - ie. ['APP_ADMIN', 'USER_ADMIN']",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: elemInSlice(validAdminRoles),
				},
			},
		},
	}
}

func resourceUserAdminRolesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	roles := convertInterfaceToStringSetNullable(d.Get("admin_roles"))
	client := getOktaClientFromMetadata(m)
	err := assignAdminRolesToUser(ctx, userId, roles, client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(userId)
	return nil
}

func resourceUserAdminRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	err := setAdminRoles(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to set read user's roles: %v", err)
	}
	return nil
}

func resourceUserAdminRolesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	client := getOktaClientFromMetadata(m)
	err := updateAdminRolesOnUser(ctx, userId, nil, client)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceUserAdminRolesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userId := d.Get("user_id").(string)
	client := getOktaClientFromMetadata(m)

	roles := convertInterfaceToStringSet(d.Get("admin_roles"))
	if err := updateAdminRolesOnUser(ctx, userId, roles, client); err != nil {
		return diag.Errorf("failed to update user admin roles: %v", err)
	}
	return nil
}
