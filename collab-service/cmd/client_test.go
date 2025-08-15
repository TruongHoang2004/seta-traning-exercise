package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"encoding/json"

	"collab-service/pkg/client"

	"github.com/stretchr/testify/assert"
)

type graphqlResponse struct {
	Data interface{} `json:"data"`
}

// create other package for test
func TestGetUsers(t *testing.T) {
	mockUsers := []client.User{
		{
			UserID:    "1",
			Username:  "Alice",
			Email:     "alice@example.com",
			Role:      client.UserTypeManager,
			CreatedAt: time.Now(),
		},
	}

	mockData := map[string]interface{}{
		"users": mockUsers,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(graphqlResponse{Data: mockData})
	}))
	defer ts.Close()

	gqlClient := client.NewGraphQLClient(ts.URL)
	userType := client.UserTypeManager
	users, err := gqlClient.GetUsers(&userType, nil)

	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Username)
}

func TestGetUser(t *testing.T) {
	mockUser := &client.User{
		UserID:    "123",
		Username:  "Bob",
		Email:     "bob@example.com",
		Role:      client.UserTypeMember,
		CreatedAt: time.Now(),
	}

	mockData := map[string]interface{}{
		"user": mockUser,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(graphqlResponse{Data: mockData})
	}))
	defer ts.Close()

	gqlClient := client.NewGraphQLClient(ts.URL)
	user, err := gqlClient.GetUser("123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Bob", user.Username)
}

func TestCreateUser(t *testing.T) {
	mockUser := &client.User{
		UserID:    "456",
		Username:  "Charlie",
		Email:     "charlie@example.com",
		Role:      client.UserTypeMember,
		CreatedAt: time.Now(),
	}

	mockData := map[string]interface{}{
		"createUser": map[string]interface{}{
			"code":    "200",
			"success": true,
			"message": "User created",
			"errors":  []string{},
			"user":    mockUser,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(graphqlResponse{Data: mockData})
	}))
	defer ts.Close()

	gqlClient := client.NewGraphQLClient(ts.URL)
	resp, err := gqlClient.CreateUser(client.CreateUserInput{
		Username: "Charlie",
		Email:    "charlie@example.com",
		Password: "secret",
		Role:     client.UserTypeMember,
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Charlie", resp.User.Username)
}
