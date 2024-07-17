package model

type Roles struct {
	Embedded struct {
		ServiceRoleDTO []struct {
			Roles []struct {
				RoleMini
				Description string `json:"description"`
			} `json:"roles"`
		} `json:"mdsServiceRoleDTOes"`
	} `json:"_embedded"`
}

type RoleMini struct {
	RoleID string `json:"roleId"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}
