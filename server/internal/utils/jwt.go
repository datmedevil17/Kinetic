package utils

import (
	"errors"

	"github.com/appwrite/sdk-for-go/v4/appwrite"
)

type AppwriteUser struct {
	ID    string
	Email string
}

var endpoint string
var projectID string

func InitAppwrite(ep, pid string) {
	endpoint = ep
	projectID = pid
}

func ValidateToken(jwtToken string) (*AppwriteUser, error) {
	if jwtToken == "" {
		return nil, errors.New("no jwt token provided")
	}

	// Create a new client instance for this specific user
	c := appwrite.NewClient(
		appwrite.WithEndpoint(endpoint),
		appwrite.WithProject(projectID),
		appwrite.WithJWT(jwtToken),
	)

	acc := appwrite.NewAccount(c)
	user, err := acc.Get()
	if err != nil {
		return nil, err
	}

	return &AppwriteUser{
		ID:    user.Id,
		Email: user.Email,
	}, nil
}
