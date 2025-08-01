package client

import (
	"context"
	"time"

	"github.com/machinebox/graphql"
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

// Client struct
type GraphQLClient struct {
	client      *graphql.Client
	accessToken string
}

func NewGraphQLClient(endpoint string) *GraphQLClient {
	return &GraphQLClient{
		client: graphql.NewClient(endpoint),
	}
}

func (c *GraphQLClient) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *GraphQLClient) doRequest(req *graphql.Request, respData interface{}) error {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	ctx := context.Background()
	return c.client.Run(ctx, req, respData)
}

func (c *GraphQLClient) GetUsers(role UserType) ([]User, error) {
	query := `query GetUsers($role: UserType!) {
		users(role: $role) {
			userId
			username
			email
			role
			createdAt
		}
	}`

	req := graphql.NewRequest(query)
	req.Var("role", role)

	var resp struct {
		Users []User `json:"users"`
	}

	err := c.doRequest(req, &resp)
	return resp.Users, err
}

func (c *GraphQLClient) GetUser(userID string) (*User, error) {
	query := `query GetUser($userId: ID!) {
		user(userId: $userId) {
			userId
			username
			email
			role
			createdAt
		}
	}`
	req := graphql.NewRequest(query)
	req.Var("userId", userID)

	var resp struct {
		User *User `json:"user"`
	}

	err := c.doRequest(req, &resp)
	return resp.User, err
}

func (c *GraphQLClient) CreateUser(input CreateUserInput) (*UserMutationResponse, error) {
	query := `mutation CreateUser($input: CreateUserInput!) {
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
	}`
	req := graphql.NewRequest(query)
	req.Var("input", input)

	var resp struct {
		CreateUser UserMutationResponse `json:"createUser"`
	}

	err := c.doRequest(req, &resp)
	return &resp.CreateUser, err
}

func (c *GraphQLClient) UpdateUser(userID, username, email string) (*UserMutationResponse, error) {
	query := `mutation UpdateUser($userId: ID!, $username: String!, $email: String!) {
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
	}`
	req := graphql.NewRequest(query)
	req.Var("userId", userID)
	req.Var("username", username)
	req.Var("email", email)

	var resp struct {
		UpdateUser UserMutationResponse `json:"updateUser"`
	}

	err := c.doRequest(req, &resp)
	return &resp.UpdateUser, err
}

func (c *GraphQLClient) Login(input UserInput) (*AuthMutationResponse, error) {
	query := `mutation Login($input: UserInput!) {
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
	}`
	req := graphql.NewRequest(query)
	req.Var("input", input)

	var resp struct {
		Login AuthMutationResponse `json:"login"`
	}

	err := c.doRequest(req, &resp)
	if err == nil && resp.Login.Success && resp.Login.AccessToken != "" {
		c.SetAccessToken(resp.Login.AccessToken)
	}
	return &resp.Login, err
}

func (c *GraphQLClient) RenewToken(refreshToken string) (*AuthMutationResponse, error) {
	query := `mutation RenewToken($refreshToken: String!) {
		r enewToken(refreshToken: $refreshToken) {
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
	}`
	req := graphql.NewRequest(query)
	req.Var("refreshToken", refreshToken)

	var resp struct {
		RenewToken AuthMutationResponse `json:"renewToken"`
	}

	err := c.doRequest(req, &resp)
	if err == nil && resp.RenewToken.Success && resp.RenewToken.AccessToken != "" {
		c.SetAccessToken(resp.RenewToken.AccessToken)
	}
	return &resp.RenewToken, err
}

func (c *GraphQLClient) ValidateToken(ctx context.Context, token string) (*User, error) {
	query := `query ParseToken($accessToken: String!) {
		parseToken(accessToken: $accessToken) {
			userId
			username
			email
			role
		}
	}`

	req := graphql.NewRequest(query)
	req.Var("accessToken", token)

	var resp struct {
		ParseToken User `json:"parseToken"`
	}

	// Use the external context in the request
	err := c.client.Run(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.ParseToken, nil
}
