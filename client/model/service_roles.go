package model

type ServiceRoles struct {
	Embedded struct {
		ServiceRoleDTO []struct {
			Roles []struct {
				RoleID      string                   `json:"roleId"`
				Name        string                   `json:"name"`
				Type        string                   `json:"type"`
				Permissions []ServiceRolePermissions `json:"permissions"`
				Description string                   `json:"description"`
			} `json:"roles"`
		} `json:"mdsServiceRoleDTOes"`
	} `json:"_embedded"`
}

type ServiceRolePermissions struct {
	Description  string `json:"description"`
	Name         string `json:"name"`
	PermissionId string `json:"permissionId"`
}
