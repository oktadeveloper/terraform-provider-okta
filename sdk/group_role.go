package sdk

import (
	"fmt"

	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
)

type (
	Role struct {
		AssignmentType string `json:"assignmentType,omitempty"`
		Id             string `json:"id,omitempty"`
		Status         string `json:"status,omitempty"`
		Type           string `json:"type,omitempty"`
	}

	RoleGroupTarget struct {
		Id string `json:"id,omitempty"`
	}
)

var ValidAdminRoles = []string{"SUPER_ADMIN", "ORG_ADMIN", "API_ACCESS_MANAGEMENT_ADMIN", "APP_ADMIN", "USER_ADMIN", "MOBILE_ADMIN", "READ_ONLY_ADMIN", "HELP_DESK_ADMIN"}
var ValidAdminRolesForGroupTarget = []string{"USER_ADMIN", "HELP_DESK_ADMIN"}

func (m *ApiSupplement) DeleteAdminRole(groupId, roleId string) (*okta.Response, error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles/%s", groupId, roleId)
	req, err := m.RequestExecutor.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return m.RequestExecutor.Do(req, nil)
}

func (m *ApiSupplement) ListAdminRoles(groupId string, qp *query.Params) (roles []*Role, resp *okta.Response, err error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles", groupId)
	if qp != nil {
		url += qp.String()
	}
	req, err := m.RequestExecutor.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	resp, err = m.RequestExecutor.Do(req, &roles)
	return
}

func (m *ApiSupplement) CreateAdminRole(groupId string, body *Role, qp *query.Params) (*Role, *okta.Response, error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles", groupId)
	if qp != nil {
		url += qp.String()
	}
	req, err := m.RequestExecutor.NewRequest("POST", url, body)
	if err != nil {
		return nil, nil, err
	}

	respBody := &Role{}
	resp, err := m.RequestExecutor.Do(req, respBody)
	return respBody, resp, err
}

func (m *ApiSupplement) DeleteAdminRoleGroupTarget(groupId, roleId, targetGroupId string) (*okta.Response, error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles/%s/targets/groups/%s", groupId, roleId, targetGroupId)
	req, err := m.RequestExecutor.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return m.RequestExecutor.Do(req, nil)
}

func (m *ApiSupplement) ListAdminRoleGroupTargets(groupId, roleId string, qp *query.Params) (rolegrouptargets []*RoleGroupTarget, resp *okta.Response, err error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles/%s/targets/groups", groupId, roleId)
	if qp != nil {
		url += qp.String()
	}
	req, err := m.RequestExecutor.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	resp, err = m.RequestExecutor.Do(req, &rolegrouptargets)
	return
}

func (m *ApiSupplement) CreateAdminRoleGroupTarget(groupId, roleId, targetGroupId string) (*okta.Response, error) {
	url := fmt.Sprintf("/api/v1/groups/%s/roles/%s/targets/groups/%s", groupId, roleId, targetGroupId)

	req, err := m.RequestExecutor.NewRequest("PUT", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.RequestExecutor.Do(req, nil)
	return resp, err
}
