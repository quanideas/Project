package projecthandlers

import (
	"fmt"
	"os"
	"project/common/constants"
	"project/common/helpers"
	commonqueries "project/common/queries"
	"project/database"
	"project/models/entity"
	"project/models/request"
	"project/models/response"

	"github.com/devfeel/mapper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create is used to create a project template. Can only be used by root.
// Params
// { Name: name of the project
// CompanyID: company project belongs to }
func Create(c *fiber.Ctx) error {
	// Parse user to match entity model
	userRequest := request.CreateProjectRequest{}
	if err := c.BodyParser(&userRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)

	// Only allow when user is root user
	if !isRoot {
		helpers.BadRequest(c,
			"No permission to create a project user",
			constants.ERR_PROJECT_CREATION_NOT_ALLOWED)
		return nil
	}

	// Fill in create metadata
	project := entity.Project{}
	mapper.Mapper(&userRequest, &project)
	project.BaseEntityModel = helpers.CreateMetaData(username)

	// Call db to create project
	result := database.DB.Db.Create(&project)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map to return model
	projectResponse := response.ProjectResponse{}
	mapper.Mapper(&project, &projectResponse)

	// Return created project
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: projectResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// Get all
func GetAll(c *fiber.Ctx) error {
	// Parse get all query model
	getAllRequest := request.GetAll{}
	if err := c.BodyParser(&getAllRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username and companyID from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	companyID := claims["company_id"].(string)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// If root then get all projects
	var query *gorm.DB
	var err error
	if isRoot {
		if getAllRequest.CompanyID == "" {
			query = GetAllRoot()
		} else {
			query = getAllByCompany(getAllRequest.CompanyID)
		}
	} else if isAdmin { // Else if admin then get projects of this user's company
		query = getAllByCompany(companyID)
	} else { // Else check permission
		query, err = getAllRegularUser(companyID, username, token, refreshToken)
	}
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Handle search and sort
	query, errCode, err := commonqueries.AddSearchAndSortGetAll(getAllRequest, query, entity.Project{})
	if err != nil {
		helpers.BadRequest(c, err.Error(), errCode)
		return nil
	}

	// Get total number
	var count int64
	query.Count(&count)

	// Limit
	query = query.Offset((getAllRequest.Page - 1) * getAllRequest.Count).Limit(getAllRequest.Count)

	// Get
	var projects []response.ProjectResponse
	result := query.Find(&projects)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return get all projects
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: response.GetAll{
			List:  projects,
			Total: count,
		},
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetByID returns full project information with project iterations.
// User needs permission to view the project to get.
// Params
// { ID: id of the project }
func GetByID(c *fiber.Ctx) error {
	// Parse get all query model
	userRequest := request.GetByIDRequest{}
	if err := c.BodyParser(&userRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}
	projectID, err := uuid.Parse(userRequest.ID)
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username and companyID from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)
	companyID := claims["company_id"].(string)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	if !isRoot && !isAdmin { // Check for permission if not root
		// Get info of User microservice
		host := os.Getenv("USER_SERVICE_HOST")
		port := os.Getenv("USER_SERVICE_PORT")
		api := constants.PermissionValidate
		url := fmt.Sprintf("%s:%s/permission%s", host, port, api)

		// Check if user has permission view to this project
		agent := fiber.Post(url)
		agent.JSON(request.GetUserSpecificPermissionRequest{
			ProjectID:       &projectID,
			PermissionType:  constants.PERM_PROJECT,
			PermissionLevel: constants.PERM_LEVEL_VIEW,
		})

		// Get permission
		var data string
		if errCode, err := helpers.SendAndParseResponseData(agent, &data, token, refreshToken); err != nil {
			helpers.BadRequest(c, err.Error(), errCode)
			return nil
		}

		// Permission denied then return denied
		if data != "Granted" {
			helpers.BadRequest(c, "no permission", constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
			return nil
		}
	}

	// Query project
	var projectResponse response.ProjectGetByIDResponse
	result := database.DB.Db.Model(entity.Project{}).Where("id = ?", projectID).First(&projectResponse)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "project not found", constants.ERR_PROJECT_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Check for project is in user's company if user is admin
	if isAdmin {
		if projectResponse.CompanyID.String() != companyID {
			helpers.BadRequest(c, "project not found", constants.ERR_PROJECT_NOT_FOUND)
			return nil
		}
	}

	// Query project iterations
	result = database.DB.Db.Model(entity.ProjectIteration{}).Where("project_id = ?", projectID).
		Find(&projectResponse.ProjectIteration)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}
	projectResponse.NumOfIterations = len(projectResponse.ProjectIteration)

	// Return get all projects
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: projectResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// GetCompanyIDByProjectID returns company ID that the project belongs to.
// Only root can call this.
// Params
// { ID: id of the project }
func GetCompanyIDByProjectID(c *fiber.Ctx) error {
	// Parse get all query model
	request := request.GetByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}
	projectID, err := uuid.Parse(request.ID)
	if err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username and companyID from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)

	// Only allow root
	if !isRoot {
		helpers.BadRequest(c, "no permission", constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
	}

	// Query company ID
	var companyID string
	result := database.DB.Db.Model(&entity.Project{}).
		Where("id = ?", projectID).
		Select("company_id").First(&companyID)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "project not found", constants.ERR_PROJECT_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return company ID
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: companyID,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}
