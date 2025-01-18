package main

import (
	"log"
	"os"
	"project/database"
	"project/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("./.env"); err != nil {
		if _, envExist := os.LookupEnv("ENV_LOADED"); !envExist {
			log.Fatal("environment file not found")
		}
	}

	database.ConnectDb()

	server.RunServer()
}
