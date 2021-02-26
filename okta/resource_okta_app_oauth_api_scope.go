package okta

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
)

func resourceAppOAuthApiScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppOAuthApiScopeCreate,
		ReadContext:   resourceAppOAuthApiScopeRead,
		UpdateContext: resourceAppOAuthApiScopeUpdate,
		DeleteContext: resourceAppOAuthApiScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				scopes, _, err := getOktaClientFromMetadata(m).Application.ListScopeConsentGrants(ctx, d.Id(), nil)
				if err != nil {
					return nil, err
				}

				_ = d.Set("app_id", d.Id())
				if len(scopes) > 0 {
					// Assume issuer is the same for all granted scopes, taking the first
					_ = d.Set("issuer", scopes[0].Issuer)
				} else {
					return nil, errors.New("no application scope found")
				}
				if err = setOAuthApiScopes(d, scopes); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "ID of the application.",
				ForceNew:    true,
			},
			"issuer": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "The issuer of your Org Authorization Server, your Org URL.",
			},
			"scopes": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Scopes of the application for which consent is granted.",
			},
		},
	}
}

func resourceAppOAuthApiScopeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	scopes := make([]string, 0)

	for _, scope := range d.Get("scopes").([]interface{}) {
		scopes = append(scopes, scope.(string))
	}
	grantScopeList := getOAuthApiScopeList(scopes, d.Get("issuer").(string))
	err := grantOAuthApiScopes(ctx, d, m, grantScopeList)
	if err != nil {
		return diag.Errorf("failed to create application scope consent grant: %v", err)
	}

	return resourceAppOAuthApiScopeRead(ctx, d, m)
}

func resourceAppOAuthApiScopeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	scopes, _, err := getOktaClientFromMetadata(m).Application.ListScopeConsentGrants(ctx, d.Get("app_id").(string), nil)
	if err != nil {
		return diag.Errorf("failed to get application scope consent grants: %v", err)
	}

	if scopes == nil {
		d.SetId("")
		return nil
	}

	err = setOAuthApiScopes(d, scopes)
	if err != nil {
		return diag.Errorf("failed to set application scope consent grant: %v", err)
	}

	return nil
}

func resourceAppOAuthApiScopeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	scopes, _, err := getOktaClientFromMetadata(m).Application.ListScopeConsentGrants(ctx, d.Get("app_id").(string), nil)

	if err != nil {
		return diag.Errorf("failed to get application scope consent grants: %v", err)
	}

	grantList, revokeList := getOAuthApiScopeUpdateLists(d, scopes)
	grantScopeList := getOAuthApiScopeList(grantList, d.Get("issuer").(string))
	err = grantOAuthApiScopes(ctx, d, m, grantScopeList)
	if err != nil {
		return diag.Errorf("failed to create application scope consent grant: %v", err)
	}

	scopeMap, err := getOAuthApiScopeIdMap(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to get application scope consent grant: %v", err)
	}

	revokeListIds := make([]string, 0)
	for _, scope := range revokeList {
		revokeListIds = append(revokeListIds, scopeMap[scope])
	}
	err = revokeOAuthApiScope(ctx, d, m, revokeListIds)
	if err != nil {
		return diag.Errorf("failed to revoke application scope consent grant: %v", err)
	}

	return resourceAppOAuthApiScopeRead(ctx, d, m)
}

func resourceAppOAuthApiScopeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	scopeMap, err := getOAuthApiScopeIdMap(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to get application scope consent grant: %v", err)
	}

	revokeListIds := make([]string, 0)
	for _, scope := range d.Get("scopes").([]interface{}) {
		revokeListIds = append(revokeListIds, scopeMap[scope.(string)])
	}
	err = revokeOAuthApiScope(ctx, d, m, revokeListIds)
	if err != nil {
		return diag.Errorf("failed to revoke application scope consent grant: %v", err)
	}

	return nil
}

// Resource Helpers
// Creates a new OAuth2ScopeConsentGrant struct
func newOAuthApiScope(scopeId string, issuer string) *okta.OAuth2ScopeConsentGrant {
	return &okta.OAuth2ScopeConsentGrant{
		Issuer:  issuer,
		ScopeId: scopeId,
	}
}

// Creates a list of OAuth2ScopeConsentGrant structs from a string list with scope names
func getOAuthApiScopeList(scopeIds []string, issuer string) []*okta.OAuth2ScopeConsentGrant {
	result := make([]*okta.OAuth2ScopeConsentGrant, len(scopeIds))
	for i, scopeId := range scopeIds {
		result[i] = newOAuthApiScope(scopeId, issuer)
	}
	return result
}

// Fetches current granted application scopes and returns a map with names and IDs.
func getOAuthApiScopeIdMap(ctx context.Context, d *schema.ResourceData, m interface{}) (map[string]string, error) {
	result := make(map[string]string)
	currentScopes, resp, err := getOktaClientFromMetadata(m).Application.ListScopeConsentGrants(ctx, d.Get("app_id").(string), nil)
	if err := suppressErrorOn404(resp, err); err != nil {
		return nil, fmt.Errorf("failed to get application scope consent grants: %v", err)
	}
	for _, currentScope := range currentScopes {
		result[currentScope.ScopeId] = currentScope.Id
	}
	return result, nil
}

// set resource schema from a list scopes
func setOAuthApiScopes(d *schema.ResourceData, to []*okta.OAuth2ScopeConsentGrant) error {
	scopes := make([]string, len(to))
	for i, scope := range to {
		scopes[i] = scope.ScopeId
	}
	d.SetId(d.Get("app_id").(string))
	_ = d.Set("issuer", d.Get("issuer").(string))
	_ = d.Set("scopes", scopes)
	return nil
}

// Grant a list of scopes to an OAuth application. For convenience this function takes a list of OAuth2ScopeConsentGrant structs.
func grantOAuthApiScopes(ctx context.Context, d *schema.ResourceData, m interface{}, scopeGrants []*okta.OAuth2ScopeConsentGrant) error {
	for _, scopeGrant := range scopeGrants {
		_, _, err := getOktaClientFromMetadata(m).Application.GrantConsentToScope(ctx, d.Get("app_id").(string), *scopeGrant)
		if err != nil {
			return fmt.Errorf("failed to grant application api scope: %v", err)
		}
	}
	return nil
}

// Revoke a list of scopes from an OAuth application. The scope ID is needed for a revoke.
func revokeOAuthApiScope(ctx context.Context, d *schema.ResourceData, m interface{}, ids []string) error {
	for _, id := range ids {
		resp, err := getOktaClientFromMetadata(m).Application.RevokeScopeConsentGrant(ctx, d.Get("app_id").(string), id)
		if err := suppressErrorOn404(resp, err); err != nil {
			return fmt.Errorf("failed to revoke application api scope: %v", err)
		}
	}
	return nil
}

// Diff function to identify which scope needs to be added or removed to the application
func getOAuthApiScopeUpdateLists(d *schema.ResourceData, from []*okta.OAuth2ScopeConsentGrant) ([]string, []string) {
	grantList := make([]string, 0)
	revokeList := make([]string, 0)
	desiredScopes := make([]string, 0)
	currentScopes := make([]string, 0)

	// cast list of interface{} to strings
	for _, scope := range d.Get("scopes").([]interface{}) {
		desiredScopes = append(desiredScopes, scope.(string))
	}

	// extract scope list form []okta.OAuth2ScopeConsentGrant
	for _, currentScope := range from {
		currentScopes = append(currentScopes, currentScope.ScopeId)
	}

	// find scopes that should not be there
	for _, currentScope := range currentScopes {
		if !contains(desiredScopes, currentScope) {
			revokeList = append(revokeList, currentScope)
		}
	}

	// scopes that need to be granted
	for _, desiredScope := range desiredScopes {
		if !contains(currentScopes, desiredScope) {
			grantList = append(grantList, desiredScope)
		}
	}

	return grantList, revokeList
}
