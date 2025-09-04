package user_service

import (
	"collab-service/internal/domain/entity"
	"collab-service/internal/infrastructure/external/user_service/queries"
	"collab-service/internal/infrastructure/logger"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/machinebox/graphql"
	"github.com/sony/gobreaker"
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
	breaker     *gobreaker.CircuitBreaker
}

func NewGraphQLClient(endpoint string) *GraphQLClient {
	cbSettings := gobreaker.Settings{
		Name:        "GraphQLClientBreaker",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
	}

	return &GraphQLClient{
		client:  graphql.NewClient(endpoint),
		breaker: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

func (c *GraphQLClient) SetAccessToken(token string) {
	c.accessToken = token
}

// Kiểm tra breaker có đang mở không
func (c *GraphQLClient) IsBreakerOpen() bool {
	return c.breaker.State() == gobreaker.StateOpen
}

func (c *GraphQLClient) doRequest(req *graphql.Request, respData interface{}) error {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	_, err := c.breaker.Execute(func() (interface{}, error) {
		runErr := c.client.Run(context.Background(), req, respData)
		if runErr != nil {
			return nil, runErr
		}

		// Nếu response có trường Code thì check
		if resp, ok := respData.(*struct {
			Code string `json:"code"`
		}); ok && resp.Code >= "500" {
			return nil, fmt.Errorf("server error: code=%s", resp.Code)
		}

		return respData, nil
	})

	if err != nil {
		if c.IsBreakerOpen() {
			logger.Error("Breaker is OPEN, request blocked", err)
		}
		return err
	}
	return nil
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

func (c *GraphQLClient) Ping(ctx context.Context) (string, error) {
	req := graphql.NewRequest(queries.Ping)

	var resp struct {
		Ping struct {
			Code    string   `json:"code"`
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Errors  []string `json:"errors"`
		} `json:"ping"`
	}

	if ctx == nil {
		ctx = context.Background()
	}

	_, err := c.breaker.Execute(func() (interface{}, error) {
		runErr := c.client.Run(ctx, req, &resp)
		if runErr != nil {
			return nil, runErr
		}

		if resp.Ping.Code >= "500" {
			return nil, fmt.Errorf("server error: %s", resp.Ping.Message)
		}
		return resp.Ping.Message, nil
	})

	if err != nil {
		if c.IsBreakerOpen() {
			logger.Error("Breaker is OPEN, Ping request blocked", err)
		} else {
			logger.Error(fmt.Sprintf("Ping failed: %v", err))
		}
		return "", err
	}

	log.Println("Ping response from user service:", resp.Ping)
	return resp.Ping.Message, nil
}

func (c *GraphQLClient) CreateUser(ctx context.Context, input CreateUserInput) (*UserMutationResponse, error) {
	req := graphql.NewRequest(queries.CreateUser)
	req.Var("input", input)
	var resp struct {
		CreateUser UserMutationResponse `json:"createUser"`
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err := c.client.Run(ctx, req, &resp)
	if resp.CreateUser.Errors != nil {
		return nil, fmt.Errorf("failed to create user: %v", resp.CreateUser.Errors)
	}
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
