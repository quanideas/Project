package projecthandlers

import (
	"fmt"
	"os"
	"project/common/constants"
	"project/common/helpers"
	"project/models/request"
	"project/models/response"
	"project/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetAllRoot() *gorm.DB {
	// Init query
	return repositories.GetAllRoot()

}

func getAllByCompany(companyID string) *gorm.DB {
	// Init query
	return repositories.GetAllByCompany(companyID)

}

func getAllRegularUser(companyID, username string, token, refreshToken string) (*gorm.DB, error) {
	var projectIDs []uuid.UUID

	// Get info of User microservice
	host := os.Getenv("USER_SERVICE_HOST")
	port := os.Getenv("USER_SERVICE_PORT")
	api := constants.UserGetSpecificPermission
	url := fmt.Sprintf("%s:%s/user/%s", host, port, api)

	// Check if user has permission VIEW_ALL_PROJECTS
	agent := fiber.Post(url)
	agent.JSON(request.GetUserSpecificPermissionRequest{
		ProjectID:       nil,
		PermissionType:  constants.PERM_VIEW_ALL_PROJECT,
		PermissionLevel: "",
	})

	// Get permission
	var data []response.UserGetSpecificPermissionResponse
	if _, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken); err != nil {
		return nil, err
	}

	// User have view_all_projects permission, then query with full permission
	if len(data) != 0 {
		return repositories.GetAllByCompany(companyID), nil
	} else { // User doesn't have permission, then get all projects this user has permission to
		// Get all projects from user_project_permissions and role_project_permissions
		agent := fiber.Post(url)
		agent.JSON(request.GetUserSpecificPermissionRequest{
			ProjectID:       nil,
			PermissionType:  "project",
			PermissionLevel: "",
		})
		agent.Set("Authorization", token)

		// Get project permission
		var data []response.UserGetSpecificPermissionResponse
		if _, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken); err != nil {
			return nil, err
		}

		// Parse project IDs to string array
		for _, permission := range data {
			projectIDs = append(projectIDs, *permission.ProjectID)
		}
	}

	// Init query
	return repositories.GetProjectListRegularUser(companyID, username, projectIDs), nil
}
