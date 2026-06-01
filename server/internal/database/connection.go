package database

import (
	"log"

	"github.com/appwrite/sdk-for-go/v4/appwrite"
	"github.com/appwrite/sdk-for-go/v4/client"
	"github.com/appwrite/sdk-for-go/v4/databases"
)

var Client client.Client
var Databases *databases.Databases

func Connect(endpoint, projectID, apiKey string) {
	Client = appwrite.NewClient(
		appwrite.WithEndpoint(endpoint),
		appwrite.WithProject(projectID),
		appwrite.WithKey(apiKey),
	)

	Databases = appwrite.NewDatabases(Client)

	log.Println("✅ Appwrite connected successfully")
}
