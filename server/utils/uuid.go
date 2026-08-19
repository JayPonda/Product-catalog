package utils

import "github.com/google/uuid"

func GetUUID() (uuid.UUID, error) {
	return uuid.NewUUID()
}
