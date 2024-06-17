package service_type

import (
	"fmt"
	"strings"
)

const (
	RABBITMQ = "RABBITMQ"
	MYSQL    = "MYSQL"
	POSTGRES = "POSTGRES"
	REDIS    = "REDIS"
)

func GetAll() []string {
	return []string{
		POSTGRES,
		MYSQL,
		RABBITMQ,
		REDIS,
	}
}

func ValidateRoleType(stateType string) error {
	switch stateType {
	case MYSQL, RABBITMQ, POSTGRES, REDIS:
		return nil
	default:
		return fmt.Errorf("invalid type: supported types are [%s]",
			strings.Join(GetAll(), ", "))
	}
}
