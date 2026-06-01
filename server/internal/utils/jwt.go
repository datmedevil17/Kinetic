package utils

import (
	"errors"
	"fmt"

	supa "github.com/supabase-community/supabase-go"
)

// SupabaseUser holds the fields we care about from Supabase's user response.
type SupabaseUser struct {
	ID    string
	Email string
}

// supabaseURL and supabaseServiceKey are stored at init time
// so ValidateToken can make raw HTTP calls without re-creating the client.
var (
	SupabaseClient     *supa.Client
	supabaseURL        string
	supabaseServiceKey string
)

// InitSupabase creates the Supabase client.
// Call this once in main.go after loading config.
func InitSupabase(url, serviceRoleKey string) {
	var err error
	SupabaseClient, err = supa.NewClient(url, serviceRoleKey, &supa.ClientOptions{})
	if err != nil {
		panic(fmt.Sprintf("failed to create Supabase client: %v", err))
	}
	supabaseURL = url
	supabaseServiceKey = serviceRoleKey
}

// ValidateToken validates a Supabase JWT by calling the GoTrue Auth API.
// The supabase-go library exposes Auth.GetUser(token) which does exactly this.
//
// Returns the authenticated user's ID and email, or an error if invalid.
func ValidateToken(accessToken string) (*SupabaseUser, error) {
	if accessToken == "" {
		return nil, errors.New("no access token provided")
	}

	// SupabaseClient.Auth is a gotrue.Client interface.
	// GetUser(token) calls GET /auth/v1/user with Bearer <token>.
	user, err := SupabaseClient.Auth.WithToken(accessToken).GetUser()
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if user.ID.String() == "" {
		return nil, errors.New("invalid user response from Supabase")
	}

	return &SupabaseUser{
		ID:    user.ID.String(),
		Email: user.Email,
	}, nil
}

