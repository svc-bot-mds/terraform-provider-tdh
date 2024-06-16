package identity_type

import (
	"fmt"
	"strings"
)

const (
	UserAccount      = "USER_ACCOUNT"
	LocalUserAccount = "LOCAL_USER_ACCOUNT"
)

func GetAll() []string {
	return []string{
		UserAccount,
		LocalUserAccount,
	}
}

func Validate(stateType string) error {
	switch stateType {
	case UserAccount, LocalUserAccount:
		return nil
	default:
		return fmt.Errorf("invalid type: supported types are [%s]",
			strings.Join(GetAll(), ", "))
	}
}
