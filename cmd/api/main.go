package main

import (
	"log"
	"os"
	"strconv"

	"city-tags-api/internal/server"

	_ "github.com/joho/godotenv/autoload"
)

//	@title			City Tags API
//	@version		0.0.3
//	@description	This is an API that makes available different tags for worlwide cities

//	@contact.name	City Tags API Support
//	@contact.email	a.jaenrev@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		city-tags-api.com
//	@BasePath	/api/v0

//	@securityDefinitions.basic	BasicAuth

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Authorization to access the API endpoints

func main() {
	envPort := os.Getenv("SERVER_PORT")
	port, err := strconv.Atoi(envPort)
	if err != nil {
		log.Fatalf("SERVER_PORT env variable error when parsing to integer %s", err.Error())
	}

	cityTApi := server.NewServer(port)
	log.Printf("Running server on %s", cityTApi.Addr)
	if err := cityTApi.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
