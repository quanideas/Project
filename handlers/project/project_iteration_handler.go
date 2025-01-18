package projecthandlers

import (
	"fmt"
	"os"
	"project/common/constants"
	"project/common/helpers"
	"project/database"
	"project/models/entity"
	"project/models/request"
	"project/models/response"

	"github.com/devfeel/mapper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// GetIterationByID returns an iteration record with permission validation.
// Param
// id: permission iteration's ID
func GetIterationByID(c *fiber.Ctx) error {
	// Parse user to match entity model
	userRequest := request.GetByIDRequest{}
	if err := c.BodyParser(&userRequest); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	userLocal := c.Locals("user").(*jwt.Token)
	claims := userLocal.Claims.(jwt.MapClaims)
	isRoot := claims["is_root"].(bool)
	isAdmin := claims["is_admin"].(bool)
	companyID := claims["company_id"].(string)
	token := c.Cookies("token")
	refreshToken := c.Cookies("refreshToken")

	// Get iteration from db
	var iteration entity.ProjectIteration
	result := database.DB.Db.Where("id = ?", userRequest.ID).First(&iteration)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "project iteration not found", constants.ERR_PROJECT_ITERATION_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Check if this iteration belongs to user's company if user is admin
	if isAdmin {
		var count int64
		result := database.DB.Db.Model(entity.Project{}).
			Where("id = ? AND company_id = ?", iteration.ProjectID, companyID).
			Count(&count)
		if result.Error != nil {
			helpers.InternalServerError(c, result.Error.Error())
			return nil
		}

		if count == 0 {
			helpers.BadRequest(c, "no permission", constants.ERR_COMMON_PERMISSION_NOT_ALLOWED)
		}
	} else if !isRoot { // Check if regular user have permission to view this project
		// Get info of User microservice
		host := os.Getenv("USER_SERVICE_HOST")
		port := os.Getenv("USER_SERVICE_PORT")
		api := constants.PermissionValidate
		url := fmt.Sprintf("%s:%s/permission%s", host, port, api)

		// Check if user has permission view to this project
		agent := fiber.Post(url)
		agent.JSON(request.GetUserSpecificPermissionRequest{
			ProjectID:       &iteration.ProjectID,
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

	// Return found iteration
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: iteration,
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}

// CreateIteration creates a project iteration record. NO file upload is handled here.
// Clients need to call file manager service to upload files.
// Can only be used by root users.
func CreateIteration(c *fiber.Ctx) error {
	// Parse user to match entity model
	userRequest := request.CreateIterationRequest{}
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
			"no permission to upload",
			constants.ERR_PROJECT_ITERATION_UPLOAD_NOT_ALLOWED)
		return nil
	}

	// Fill in create metadata
	iteration := entity.ProjectIteration{}
	mapper.Mapper(&userRequest, &iteration)
	iteration.BaseEntityModel = helpers.CreateMetaData(username)

	// Call db to create iteration
	result := database.DB.Db.Create(&iteration)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Map to return model
	iterationResponse := response.IterationResponse{}
	mapper.Mapper(&iteration, &iterationResponse)

	// Return created iteration
	c.Status(201)
	c.JSON(response.BaseResponse{
		Data: iterationResponse,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// Updates returns updated project iteration's information if successful.
// Can only be used by root users.
// Params takes in full project iteration info
func Update(c *fiber.Ctx) error {
	// Parse Project Iteration to match entity model
	request := entity.ProjectIteration{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Get username from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	isRoot := claims["is_root"].(bool)

	// Only allow when user is root user
	if !isRoot {
		helpers.BadRequest(c,
			"no permission to upload",
			constants.ERR_PROJECT_ITERATION_UPLOAD_NOT_ALLOWED)
		return nil
	}

	// Check if project iteration exists on db
	var iteration entity.ProjectIteration
	result := database.DB.Db.Model(&entity.ProjectIteration{}).Where("id = ?", request.ID).First(&iteration)
	if result.Error == gorm.ErrRecordNotFound {
		helpers.BadRequest(c, "project iteration not found", constants.ERR_PROJECT_ITERATION_NOT_FOUND)
		return nil
	} else if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Fill in metadata
	request.BaseEntityModel = helpers.EditMetaData(username, iteration.BaseEntityModel)
	request.ProjectID = iteration.ProjectID

	// Update
	result = database.DB.Db.Save(&request)
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return edited iteration
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: request,
		Meta: struct{ Status int }{Status: 200},
	})

	return nil
}

// Delete deletes an iteration from a project.
// This only removes a record on db, not deleting the actual files.
// Param takes in the ID of the iteration.
func Delete(c *fiber.Ctx) error {
	// Parse request model
	request := request.DeleteByIDRequest{}
	if err := c.BodyParser(&request); err != nil {
		helpers.InternalServerError(c, err.Error())
		return nil
	}

	// Call db to delete
	result := database.DB.Db.Where("id = ?", request.ID).Delete(entity.ProjectIteration{})
	if result.Error != nil {
		helpers.InternalServerError(c, result.Error.Error())
		return nil
	}

	// Return success as deleted
	c.Status(200)
	c.JSON(response.BaseResponse{
		Data: "Success",
		Meta: struct{ Status int }{Status: 200},
	})
	return nil
}
