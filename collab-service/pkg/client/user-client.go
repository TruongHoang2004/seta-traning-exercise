package client

import (
	"collab-service/pkg/client/queries"
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

// Methods
func (c *GraphQLClient) GetUsers(role *UserType, userIDs []string) ([]User, error) {
	req := graphql.NewRequest(queries.GetUsers)
	if role != nil {
		req.Var("role", *role)
	}
	if len(userIDs) > 0 {
		req.Var("userIds", userIDs)
	}
	var resp struct {
		Users []User `json:"users"`
	}
	err := c.doRequest(req, &resp)
	return resp.Users, err
}

func (c *GraphQLClient) GetUser(userID string) (*User, error) {
	req := graphql.NewRequest(queries.GetUser)
	req.Var("userId", userID)
	var resp struct {
		User *User `json:"user"`
	}
	err := c.doRequest(req, &resp)
	return resp.User, err
}

func (c *GraphQLClient) CreateUser(input CreateUserInput) (*UserMutationResponse, error) {
	req := graphql.NewRequest(queries.CreateUser)
	req.Var("input", input)
	var resp struct {
		CreateUser UserMutationResponse `json:"createUser"`
	}
	err := c.doRequest(req, &resp)
	return &resp.CreateUser, err
}

func (c *GraphQLClient) UpdateUser(userID, username, email string) (*UserMutationResponse, error) {
	req := graphql.NewRequest(queries.UpdateUser)
	req.Var("userId", userID)
	req.Var("username", username)
	req.Var("email", email)
	var resp struct {
		UpdateUser UserMutationResponse `json:"updateUser"`
	}
	err := c.doRequest(req, &resp)
	return &resp.UpdateUser, err
}

func (c *GraphQLClient) ValidateToken(ctx context.Context, token string) (*User, error) {
	req := graphql.NewRequest(queries.ValidateToken)
	req.Var("accessToken", token)
	var resp struct {
		ParseToken User `json:"parseToken"`
	}
	err := c.client.Run(ctx, req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.ParseToken, nil
}
