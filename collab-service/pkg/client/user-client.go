package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Enums
type UserType string

const (
	UserTypeManager UserType = "MANAGER"
	UserTypeMember  UserType = "MEMBER"
)

// Input types
type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserInput struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserType `json:"role"`
}

// Response types
type User struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      UserType  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type Member struct {
	MemberID   string `json:"memberId"`
	MemberName string `json:"memberName"`
}

type Manager struct {
	ManagerID   string `json:"managerId"`
	ManagerName string `json:"managerName"`
}

type MutationResponse struct {
	Code    string   `json:"code"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type UserMutationResponse struct {
	MutationResponse
	User *User `json:"user"`
}

type AuthMutationResponse struct {
	MutationResponse
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         *User  `json:"user"`
}

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string        `json:"message"`
		Path    []interface{} `json:"path"`
	} `json:"errors"`
}

// Client struct
type GraphQLClient struct {
	endpoint    string
	httpClient  *http.Client
	accessToken string
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient(endpoint string) *GraphQLClient {
	return &GraphQLClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAccessToken sets the access token for authenticated requests
func (c *GraphQLClient) SetAccessToken(token string) {
	c.accessToken = token
}

// makeRequest sends a GraphQL request and returns the response
func (c *GraphQLClient) makeRequest(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(body, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return &graphqlResp, fmt.Errorf("GraphQL errors: %+v", graphqlResp.Errors)
	}

	return &graphqlResp, nil
}

// Query methods

// GetUsers retrieves users by role
func (c *GraphQLClient) GetUsers(role UserType) ([]User, error) {
	query := `
		query GetUsers($role: UserType!) {
			users(role: $role) {
				userId
				username
				email
				role
				createdAt
			}
		}
	`

	variables := map[string]interface{}{
		"role": role,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Users []User `json:"users"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return result.Users, nil
}

// GetUser retrieves a single user by ID
func (c *GraphQLClient) GetUser(userID string) (*User, error) {
	query := `
		query GetUser($userId: ID!) {
			user(userId: $userId) {
				userId
				username
				email
				role
				createdAt
			}
		}
	`

	variables := map[string]interface{}{
		"userId": userID,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		User *User `json:"user"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return result.User, nil
}

// Mutation methods

// CreateUser creates a new user
func (c *GraphQLClient) CreateUser(input CreateUserInput) (*UserMutationResponse, error) {
	query := `
		mutation CreateUser($input: CreateUserInput!) {
			createUser(input: $input) {
				code
				success
				message
				errors
				user {
					userId
					username
					email
					role
					createdAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		CreateUser UserMutationResponse `json:"createUser"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create user response: %w", err)
	}

	return &result.CreateUser, nil
}

// UpdateUser updates an existing user
func (c *GraphQLClient) UpdateUser(userID, username, email string) (*UserMutationResponse, error) {
	query := `
		mutation UpdateUser($userId: ID!, $username: String!, $email: String!) {
			updateUser(userId: $userId, username: $username, email: $email) {
				code
				success
				message
				errors
				user {
					userId
					username
					email
					role
					createdAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"userId":   userID,
		"username": username,
		"email":    email,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		UpdateUser UserMutationResponse `json:"updateUser"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal update user response: %w", err)
	}

	return &result.UpdateUser, nil
}

// Login authenticates a user
func (c *GraphQLClient) Login(input UserInput) (*AuthMutationResponse, error) {
	query := `
		mutation Login($input: UserInput!) {
			login(input: $input) {
				code
				success
				message
				errors
				accessToken
				refreshToken
				user {
					userId
					username
					email
					role
					createdAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Login AuthMutationResponse `json:"login"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	// Automatically set access token if login successful
	if result.Login.Success && result.Login.AccessToken != "" {
		c.SetAccessToken(result.Login.AccessToken)
	}

	return &result.Login, nil
}

// RenewToken refreshes the access token
func (c *GraphQLClient) RenewToken(refreshToken string) (*AuthMutationResponse, error) {
	query := `
		mutation RenewToken($refreshToken: String!) {
			renewToken(refreshToken: $refreshToken) {
				code
				success
				message
				errors
				accessToken
				refreshToken
				user {
					userId
					username
					email
					role
					createdAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"refreshToken": refreshToken,
	}

	resp, err := c.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		RenewToken AuthMutationResponse `json:"renewToken"`
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal renew token response: %w", err)
	}

	// Automatically update access token if renewal successful
	if result.RenewToken.Success && result.RenewToken.AccessToken != "" {
		c.SetAccessToken(result.RenewToken.AccessToken)
	}

	return &result.RenewToken, nil
}
