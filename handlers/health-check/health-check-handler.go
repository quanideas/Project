package healthcheck

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func ConnectionCheck(c *fiber.Ctx) error {
	healthCheck := struct {
		FileService  bool `json:"file_service"`
		UserService  bool `json:"user_service"`
		TokenService bool `json:"token_service"`
	}{}

	// Check User service
	host := os.Getenv("USER_SERVICE_HOST")
	port := os.Getenv("USER_SERVICE_PORT")
	api := "/health-check"
	url := fmt.Sprintf("%s:%s/%s", host, port, api)

	agent := fiber.Get(url)
	userStatusCode, _, _ := agent.Bytes()
	healthCheck.UserService = userStatusCode == fiber.StatusOK

	// Check File service
	host = os.Getenv("FILE_SERVICE_HOST")
	port = os.Getenv("FILE_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	fileStatusCode, _, _ := agent.Bytes()
	healthCheck.FileService = fileStatusCode == fiber.StatusOK

	// Check Token service
	host = os.Getenv("TOKEN_SERVICE_HOST")
	port = os.Getenv("TOKEN_SERVICE_PORT")
	api = "/health-check"
	url = fmt.Sprintf("%s:%s/%s", host, port, api)

	agent = fiber.Get(url)
	tokenStatusCode, _, _ := agent.Bytes()
	healthCheck.TokenService = tokenStatusCode == fiber.StatusOK

	c.JSON(healthCheck)
	return nil
}
