package valueobject

import "errors"

type Role string

const (
	RoleSystemAdmin  Role = "system_admin"
	RoleCompanyAdmin Role = "company_admin"
	RoleEmployee     Role = "employee"
	RoleAuditor      Role = "auditor"
)

var validRoles = map[Role]bool{
	RoleSystemAdmin:  true,
	RoleCompanyAdmin: true,
	RoleEmployee:     true,
	RoleAuditor:      true,
}

func NewRole(value string) (Role, error) {
	r := Role(value)
	if !validRoles[r] {
		return "", errors.New("invalid role: " + value)
	}
	return r, nil
}

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return validRoles[r]
}
