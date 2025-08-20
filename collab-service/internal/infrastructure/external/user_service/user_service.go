package user_service

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/user_service/queries"
	"collab-service/internal/infrastructure/logger"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/machinebox/graphql"
)

// Input types
type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserInput struct {
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Role     entity.UserType `json:"role"`
}

type MutationResponse struct {
	Code    string   `json:"code"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type UserMutationResponse struct {
	MutationResponse
	User *UserModel `json:"user"`
}

type AuthMutationResponse struct {
	MutationResponse
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	User         *UserModel `json:"user"`
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
func (c *GraphQLClient) GetUsers(ctx context.Context, role *entity.UserType, userIDs []string) ([]UserModel, error) {
	req := graphql.NewRequest(queries.GetUsers)
	if role != nil {
		req.Var("role", *role)
	}
	if len(userIDs) > 0 {
		req.Var("userIds", userIDs)
	}
	var resp struct {
		Users []UserModel `json:"users"`
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err := c.client.Run(ctx, req, &resp)
	return resp.Users, err
}

func (c *GraphQLClient) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	req := graphql.NewRequest(queries.GetUser)
	req.Var("userId", userID)
	var resp struct {
		User *UserModel `json:"user"`
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err := c.client.Run(ctx, req, &resp)
	return resp.User.ToDomain(), err
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

func (c *GraphQLClient) ValidateToken(ctx context.Context, token string) (*entity.User, error) {
	req := graphql.NewRequest(queries.ValidateToken)
	req.Var("accessToken", token)
	var resp struct {
		ParseToken UserModel `json:"parseToken"`
	}
	err := c.client.Run(ctx, req, &resp)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to validate token: %v", err))
		return nil, err
	}
	return resp.ParseToken.ToDomain(), nil
}
