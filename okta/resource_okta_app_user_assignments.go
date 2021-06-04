package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

func resourceAppUserAssignments() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppUserAssignmentsCreate,
		ReadContext:   resourceAppUserAssignmentsRead,
		UpdateContext: resourceAppUserAssignmentsUpdate,
		DeleteContext: resourceAppUserAssignmentsDelete,
		Importer:      nil,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "App to associate user with",
			},
			"users": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Set of users to associate with the app",
				MinItems:    1,
				Set: func(v interface{}) int {
					buf := bytes.NewBuffer(nil)
					user := v.(map[string]interface{})

					buf.WriteString(fmt.Sprintf("%s-", user["id"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", user["username"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", normalizeDataJSON(user["profile"])))

					return schema.HashString(buf.String())
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User associated with the application",
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Optional:  true,
						},
						"profile": {
							Type:             schema.TypeString,
							ValidateDiagFunc: stringIsJSON,
							StateFunc:        normalizeDataJSON,
							Optional:         true,
						},
						"retain_assignment": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Retain the user assignment on destroy. If set to true, the resource will be removed from state but not from the Okta app.",
						},
					},
				},
			},
		},
	}
}

func resourceAppUserAssignmentsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	users := d.Get("users").(*schema.Set).List()
	client := getOktaClientFromMetadata(m)
	appID := d.Get("app_id").(string)

	assignments := tfUsersToUserAssignments(users...)

	err := addUserAssignments(ctx, client, appID, assignments)
	if err != nil {
		return diag.FromErr(err)
	}

	//okta_app_user_assignments completely controls all assignments for an application
	d.SetId(appID)
	return resourceAppUserAssignmentsRead(ctx, d, m)
}

func resourceAppUserAssignmentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	appID := d.Get("app_id").(string)

	assignments, err := listApplicationUsers(ctx, client, appID)
	if err != nil {
		return diag.Errorf("failed to get users assigned to app (%s): %s", appID, err)
	}

	var tfFlattenedAssignments []interface{}

	for _, assignment := range assignments {
		tfAssignment, err := userAssignmentToTFUser(assignment)
		if err != nil {
			return diag.Errorf("failed to marshall user profile: %s", err)
		}
		tfFlattenedAssignments = append(tfFlattenedAssignments, tfAssignment)
	}

	err = d.Set("users", tfFlattenedAssignments)
	if err != nil {
		return diag.Errorf("failed to set users in tf state: %s", err)
	}
	return nil
}

func resourceAppUserAssignmentsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	appID := d.Get("app_id").(string)
	users := d.Get("users").(*schema.Set).List()

	assignments := tfUsersToUserAssignments(users...)

	err := removeUserAssignments(ctx, client, appID, assignments)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAppUserAssignmentsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getOktaClientFromMetadata(m)
	appID := d.Get("app_id").(string)

	old, new := d.GetChange("users")
	oldSet := old.(*schema.Set)
	newSet := new.(*schema.Set)

	toAdd := tfUsersToUserAssignments(
		newSet.Difference(oldSet).List()...,
	)
	toRemove := tfUsersToUserAssignments(
		oldSet.Difference(newSet).List()...,
	)

	err := removeUserAssignments(ctx, client, appID, toRemove)
	if err != nil {
		return diag.FromErr(err)
	}

	err = addUserAssignments(ctx, client, appID, toAdd)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceAppUserAssignmentsRead(ctx, d, m)
}

func tfUsersToUserAssignments(users ...interface{}) []okta.AppUser {
	var assignments []okta.AppUser

	for _, rawUser := range users {
		user := rawUser.(map[string]interface{})

		id := user["id"].(string)
		if id == "" {
			continue
		}

		username := user["username"].(string)

		password := user["password"].(string)

		rawProfile := user["profile"]
		var profile interface{}
		_ = json.Unmarshal([]byte(rawProfile.(string)), &profile)

		assignments = append(assignments, okta.AppUser{
			Id:      id,
			Profile: profile,
			Credentials: &okta.AppUserCredentials{
				UserName: username,
				Password: &okta.AppUserPasswordCredential{
					Value: password,
				},
			},
		})
	}
	return assignments
}

func userAssignmentToTFUser(assignment *okta.AppUser) (map[string]interface{}, error) {
	profile := "{}"

	jsonProfile, err := json.Marshal(assignment.Profile)
	if err != nil {
		return nil, err
	}

	if string(jsonProfile) != "" {
		profile = string(jsonProfile)
	}

	tfAssignment := map[string]interface{}{
		"id":       assignment.Id,
		"username": assignment.Credentials.UserName,
		"profile":  profile,
	}
	return tfAssignment, nil
}

func addUserAssignments(ctx context.Context, client *okta.Client, appID string, assignments []okta.AppUser) error {
	for _, assignment := range assignments {
		_, _, err := client.Application.AssignUserToApplication(ctx, appID, assignment)
		if err != nil {
			return fmt.Errorf("failed to assign user (%s) to app (%s): %s", assignment.Id, appID, err)
		}
	}
	return nil
}

func removeUserAssignments(ctx context.Context, client *okta.Client, appID string, assignments []okta.AppUser) error {
	for _, assignment := range assignments {
		_, err := client.Application.DeleteApplicationUser(ctx, appID, assignment.Id, &query.Params{})
		if err != nil {
			return fmt.Errorf("failed to unassign user (%s) from app (%s): %s", assignment.Id, appID, err)
		}
	}
	return nil
}
