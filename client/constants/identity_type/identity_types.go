package identity_type

import (
	"fmt"
	"strings"
)

const (
	UserAccount      = "USER_ACCOUNT"
	ServiceAccount   = "SERVICE_ACCOUNT"
	LocalUserAccount = "LOCAL_USER_ACCOUNT"
)

func GetAll() []string {
	return []string{
		UserAccount,
		ServiceAccount,
		LocalUserAccount,
	}
}

func Validate(stateType string) error {
	switch stateType {
	case UserAccount, LocalUserAccount, ServiceAccount:
		return nil
	default:
		return fmt.Errorf("invalid type: supported types are [%s]",
			strings.Join(GetAll(), ", "))
	}
}
