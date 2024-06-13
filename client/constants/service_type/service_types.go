package service_type

import "fmt"

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
		return fmt.Errorf("invalid type: supported types are [%s, %s, %s, %s]",
			MYSQL,
			RABBITMQ,
			POSTGRES,
			REDIS)
	}
}
