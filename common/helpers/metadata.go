package helpers

import (
	"encoding/json"
	"errors"
	"project/models/entity"
	"project/models/response"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateMetaData(username string) entity.BaseEntityModel {
	t := time.Now()
	return entity.BaseEntityModel{
		ID:           uuid.New(),
		CreatedBy:    username,
		CreatedTime:  t,
		ModifiedBy:   &username,
		ModifiedTime: &t,
	}
}

// func EditMetaData(username, createdBy string, createdTime time.Time) entity.BaseEntityModel {
// 	t := time.Now()
// 	baseEntityModel := entity.BaseEntityModel{
// 		ModifiedBy:   &username,
// 		ModifiedTime: &t,
// 	}
// 	baseEntityModel.CreatedBy = createdBy
// 	baseEntityModel.CreatedTime = createdTime

// 	return baseEntityModel
// }

func EditMetaData(username string, baseEntityFromDB entity.BaseEntityModel) entity.BaseEntityModel {
	t := time.Now()

	baseEntityFromDB.ModifiedBy = &username
	baseEntityFromDB.ModifiedTime = &t

	return baseEntityFromDB
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CompareHashedPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func SendAndParseResponseData(agent *fiber.Agent, object any, token, refreshToken string) (int, error) {
	// Validate object to be a pointer
	reflectObject := reflect.ValueOf(object)
	if reflectObject.Kind() != reflect.Pointer || reflectObject.IsNil() {
		return 500, errors.New("object is not a pointer")
	}

	// Set token and refreshToken in cookies
	agent.Cookie("token", token)
	agent.Cookie("refreshToken", refreshToken)

	// Send request
	statusCode, body, errs := agent.Bytes()

	// Handle failed request
	if len(errs) > 0 || statusCode != 200 {
		var errMsg response.ErrorResponse
		json.Unmarshal(body, &errMsg)

		return errMsg.ErrorCode, errors.New(errMsg.Error)
	}

	// Success, decode json to response DTO
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return 500, err
	}

	// Encode Data to JSON bytes, then decode to get data as object
	data := response["Data"]
	if dataEncoded, err := json.Marshal(data); err != nil {
		return 500, err
	} else if err = json.Unmarshal(dataEncoded, object); err != nil {
		return 500, err
	}

	return 0, nil
}
