package okta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

// All profile properties here so we can do a diff against the config to see if any have changed before making the
// request or before erring due to an update on a user that is DEPROVISIONED. Since we have core user props coupled
// with group/user membership a few change requests go out in the Update function.
var profileKeys = []string{
	"city",
	"cost_center",
	"country_code",
	"custom_profile_attributes",
	"department",
	"display_name",
	"division",
	"email",
	"employee_number",
	"first_name",
	"honorific_prefix",
	"honorific_suffix",
	"last_name",
	"locale",
	"login",
	"manager",
	"manager_id",
	"middle_name",
	"mobile_phone",
	"nick_name",
	"organization",
	"postal_address",
	"preferred_language",
	"primary_phone",
	"profile_url",
	"second_email",
	"state",
	"street_address",
	"timezone",
	"title",
	"user_type",
	"zip_code",
	"password",
	"recovery_question",
	"recovery_answer",
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// Supporting id and email based imports
				client := getOktaClientFromMetadata(m)
				user, _, err := client.User.GetUser(ctx, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(user.Id)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"admin_roles": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "User Okta admin roles - ie. ['APP_ADMIN', 'USER_ADMIN']",
				Deprecated:  "The `admin_roles` field is now deprecated for the resource `okta_user`, please replace all uses of this with: `okta_user_admin_roles`",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: elemInSlice(validAdminRoles),
				},
			},
			"city": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User city",
			},
			"cost_center": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User cost center",
			},
			"country_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User country code",
			},
			"custom_profile_attributes": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: stringIsJSON,
				StateFunc:        normalizeDataJSON,
				Description:      "JSON formatted custom attributes for a user. It must be JSON due to various types Okta allows.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == ""
				},
			},
			"department": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User department",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User display name, suitable to show end users",
			},
			"division": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User division",
			},
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "User primary email address",
				ValidateDiagFunc: stringIsEmail,
			},
			"employee_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User employee number",
			},
			"first_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User first name",
			},
			"group_memberships": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The groups that you want this user to be a part of. This can also be done via the group using the `users` property.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Deprecated:  "The `group_memberships` field is now deprecated for the resource `okta_user`, please replace all uses of this with: `okta_user_group_memberships`",
			},
			"honorific_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User honorific prefix",
			},
			"honorific_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User honorific suffix",
			},
			"last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User last name",
			},
			"locale": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User default location",
			},
			"login": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User Okta login",
			},
			"manager": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Manager of User",
			},
			"manager_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Manager ID of User",
			},
			"middle_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User middle name",
			},
			"mobile_phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User mobile phone number",
			},
			"nick_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User nickname",
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User organization",
			},
			"postal_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User mailing address",
			},
			"preferred_language": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User preferred language",
			},
			"primary_phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User primary phone number",
			},
			"profile_url": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "User online profile (web page)",
				ValidateDiagFunc: stringIsURL(validURLSchemes...),
			},
			"second_email": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "User secondary email address, used for account recovery",
				ValidateDiagFunc: stringIsEmail,
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User state or region",
			},
			"status": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The status of the User in Okta - remove to set user back to active/provisioned",
				Default:          statusActive,
				ValidateDiagFunc: elemInSlice([]string{statusActive, userStatusStaged, userStatusDeprovisioned, userStatusSuspended}),
				// ignore diff changing to ACTIVE if state is set to PROVISIONED or PASSWORD_EXPIRED
				// since this is a similar status in Okta terms
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == userStatusProvisioned && new == statusActive || old == userStatusPasswordExpired && new == statusActive
				},
			},
			"raw_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The raw status of the User in Okta - (status is mapped)",
			},
			"street_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User street address",
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User default timezone",
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User title",
			},
			"user_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User employee type",
			},
			"zip_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User zipcode or postal code",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "User Password",
			},
			"recovery_question": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User Password Recovery Question",
			},
			"recovery_answer": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				ValidateDiagFunc: stringLenBetween(4, 1000),
				Description:      "User Password Recovery Answer",
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logger(m).Info("creating user", "login", d.Get("login").(string))
	profile := populateUserProfile(d)
	qp := query.NewQueryParams()

	// setting activate to false on user creation will leave the user with a status of STAGED
	if d.Get("status").(string) == userStatusStaged {
		qp = query.NewQueryParams(query.WithActivate(false))
	}

	uc := &okta.UserCredentials{
		Password: &okta.PasswordCredential{
			Value: d.Get("password").(string),
		},
	}
	recoveryQuestion := d.Get("recovery_question").(string)
	recoveryAnswer := d.Get("recovery_answer").(string)
	if recoveryQuestion != "" {
		uc.RecoveryQuestion = &okta.RecoveryQuestionCredential{
			Question: recoveryQuestion,
			Answer:   recoveryAnswer,
		}
	}

	userBody := okta.CreateUserRequest{
		Profile:     profile,
		Credentials: uc,
	}
	client := getOktaClientFromMetadata(m)
	user, _, err := client.User.CreateUser(ctx, userBody, qp)
	if err != nil {
		return diag.Errorf("failed to create user: %v", err)
	}
	// set the user id into state before setting roles and status in case they fail
	d.SetId(user.Id)

	// role assigning can only happen after the user is created so order matters here
	// Only sync admin roles when it is set by a consumer as field is deprecated
	if _, exists := d.GetOk("admin_roles"); exists {
		roles := convertInterfaceToStringSetNullable(d.Get("admin_roles"))
		if roles != nil {
			err = assignAdminRolesToUser(ctx, user.Id, roles, client)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Only sync when there is opt in, consumers can chose which route they want to take
	if _, exists := d.GetOk("group_memberships"); exists {
		groups := convertInterfaceToStringSetNullable(d.Get("group_memberships"))
		err = assignGroupsToUser(ctx, user.Id, groups, client)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// status changing can only happen after user is created as well
	if d.Get("status").(string) == userStatusSuspended || d.Get("status").(string) == userStatusDeprovisioned {
		err := updateUserStatus(ctx, user.Id, d.Get("status").(string), client)
		if err != nil {
			return diag.Errorf("failed to update user status: %v", err)
		}
	}
	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logger(m).Info("reading user", "id", d.Id())
	client := getOktaClientFromMetadata(m)
	user, resp, err := client.User.GetUser(ctx, d.Id())
	if err := suppressErrorOn404(resp, err); err != nil {
		return diag.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		d.SetId("")
		return nil
	}
	_ = d.Set("raw_status", user.Status)
	rawMap := flattenUser(user)
	err = setNonPrimitives(d, rawMap)
	if err != nil {
		return diag.Errorf("failed to set user's properties: %v", err)
	}
	err = setAdminRoles(ctx, d, m)
	if err != nil {
		return diag.Errorf("failed to set user's roles: %v", err)
	}
	// Only sync when it is outlined, an empty list will remove all membership
	if _, exists := d.GetOk("group_memberships"); exists {
		err = setGroups(ctx, d, client)
		if err != nil {
			return diag.Errorf("failed to set user's groups: %v", err)
		}
	}
	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logger(m).Info("updating user", "id", d.Id())
	status := d.Get("status").(string)
	statusChange := d.HasChange("status")

	if status == userStatusStaged && statusChange {
		return diag.Errorf("Okta will not allow a user to be updated to STAGED. Can set to STAGED on user creation only")
	}

	// There are a few requests here so just making sure the state gets updated per successful downstream change
	roleChange := d.HasChange("admin_roles")
	groupChange := d.HasChange("group_memberships")
	userChange := hasProfileChange(d)
	passwordChange := d.HasChange("password")
	recoveryQuestionChange := d.HasChange("recovery_question")
	recoveryAnswerChange := d.HasChange("recovery_answer")

	// run the update status func first so a user that was previously deprovisioned
	// can be updated further if it's status changed in it's terraform configs
	client := getOktaClientFromMetadata(m)
	if statusChange {
		err := updateUserStatus(ctx, d.Id(), status, client)
		if err != nil {
			return diag.Errorf("failed to update user status: %v", err)
		}
		_ = d.Set("status", status)
	}

	if status == userStatusDeprovisioned && userChange {
		return diag.Errorf("Only the status of a DEPROVISIONED user can be updated, we detected other change")
	}

	if userChange {
		profile := populateUserProfile(d)
		userBody := okta.User{Profile: profile}
		_, _, err := client.User.UpdateUser(ctx, d.Id(), userBody, nil)
		if err != nil {
			return diag.Errorf("failed to update user: %v", err)
		}
	}

	if roleChange {
		roles := convertInterfaceToStringSet(d.Get("admin_roles"))
		if err := updateAdminRolesOnUser(ctx, d.Id(), roles, client); err != nil {
			return diag.Errorf("failed to update user: %v", err)
		}
		_ = d.Set("admin_roles", roles)
	}

	if groupChange {
		oldGM, newGM := d.GetChange("group_memberships")
		oldSet := oldGM.(*schema.Set)
		newSet := newGM.(*schema.Set)
		groupsToAdd := convertInterfaceArrToStringArr(newSet.Difference(oldSet).List())
		groupsToRemove := convertInterfaceArrToStringArr(oldSet.Difference(newSet).List())
		err := addUserToGroups(ctx, client, d.Id(), groupsToAdd)
		if err != nil {
			return diag.FromErr(err)
		}
		err = removeUserFromGroups(ctx, client, d.Id(), groupsToRemove)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if passwordChange {
		oldPassword, newPassword := d.GetChange("password")
		op := &okta.PasswordCredential{
			Value: oldPassword.(string),
		}
		np := &okta.PasswordCredential{
			Value: newPassword.(string),
		}
		npr := &okta.ChangePasswordRequest{
			OldPassword: op,
			NewPassword: np,
		}
		_, _, err := client.User.ChangePassword(ctx, d.Id(), *npr, nil)
		if err != nil {
			return diag.Errorf("failed to update user's password: %v", err)
		}
	}

	if recoveryQuestionChange || recoveryAnswerChange {
		nuc := &okta.UserCredentials{
			Password: &okta.PasswordCredential{
				Value: d.Get("password").(string),
			},
			RecoveryQuestion: &okta.RecoveryQuestionCredential{
				Question: d.Get("recovery_question").(string),
				Answer:   d.Get("recovery_answer").(string),
			},
		}
		_, _, err := client.User.ChangeRecoveryQuestion(ctx, d.Id(), *nuc)
		if err != nil {
			return diag.Errorf("failed to change user's password recovery question: %v", err)
		}
	}
	return resourceUserRead(ctx, d, m)
}

// Checks whether any profile keys have changed, this is necessary since the profile is not nested. Also, necessary
// to give a sensible user readable error when they attempt to update a DEPROVISIONED user. Previously
// this error always occurred when you set a user's status to DEPROVISIONED.
func hasProfileChange(d *schema.ResourceData) bool {
	for _, k := range profileKeys {
		if d.HasChange(k) {
			return true
		}
	}
	return false
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logger(m).Info("deleting user", "id", d.Id())
	err := ensureUserDelete(ctx, d.Id(), d.Get("status").(string), getOktaClientFromMetadata(m))
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func ensureUserDelete(ctx context.Context, id, status string, client *okta.Client) error {
	// only deprovisioned users can be deleted fully from okta
	// make two passes on the user if they aren't deprovisioned already to deprovision them first
	passes := 2
	if status == userStatusDeprovisioned {
		passes = 1
	}
	for i := 0; i < passes; i++ {
		_, err := client.User.DeactivateOrDeleteUser(ctx, id, nil)
		if err != nil {
			return fmt.Errorf("failed to deprovision or delete user from Okta: %v", err)
		}
	}
	return nil
}

func mapStatus(currentStatus string) string {
	// PASSWORD_EXPIRED and RECOVERY are effectively ACTIVE for our purposes
	if currentStatus == userStatusPasswordExpired || currentStatus == userStatusRecovery {
		return statusActive
	}
	return currentStatus
}
