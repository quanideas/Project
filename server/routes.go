package server

import (
	"os"
	"project/common/constants"
	healthcheck "project/handlers/health-check"
	projecthandlers "project/handlers/project"
	"project/server/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupRoutes(app *fiber.App) {
	allowedDevOrigins := os.Getenv("ALLOWED_DEV_ORIGINS")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	// Apply CORS
	if os.Getenv("ENVIRONMENT") == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedDevOrigins,
			AllowCredentials: true,
		}))
	} else {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowCredentials: true,
		}))
	}

	// Recover from panics
	app.Use(middlewares.CatchPanic())

	// Unauthenticated
	app.Get("/health-check", healthcheck.HealthCheck)
	app.Get("/connection-check", healthcheck.ConnectionCheck)

	// JWT Middleware
	app.Use(middlewares.ValidateJWT())

	// Authenticated
	// Project
	projectRoutes := app.Group("/project")
	projectRoutes.Post(constants.ProjectGetByID, projecthandlers.GetByID)
	projectRoutes.Post(constants.ProjectGetCompanyIDByProjectID, projecthandlers.GetCompanyIDByProjectID)
	projectRoutes.Post(constants.ProjectCreate, projecthandlers.Create)
	projectRoutes.Post(constants.ProjectGetAll, projecthandlers.GetAll)
	projectRoutes.Post(constants.ProjectIterationGet, projecthandlers.GetIterationByID)
	projectRoutes.Post(constants.ProjectIterationCreate, projecthandlers.CreateIteration)
	projectRoutes.Post(constants.ProjectIterationUpdate, projecthandlers.Update)
	projectRoutes.Post(constants.ProjectIterationDelete, projecthandlers.Delete)
}
