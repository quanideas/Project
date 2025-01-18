package response

import "github.com/google/uuid"

type UserGetSpecificPermissionResponse struct {
	ProjectID       *uuid.UUID `json:"project_id"`
	PermissionType  string     `json:"permission_type"`
	PermissionLevel string     `json:"permission_level"`
}
