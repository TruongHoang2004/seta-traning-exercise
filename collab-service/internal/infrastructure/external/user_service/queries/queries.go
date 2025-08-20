package queries

import _ "embed"

//go:embed get_users.graphql
var GetUsers string

//go:embed get_user.graphql
var GetUser string

//go:embed create_user.graphql
var CreateUser string

//go:embed update_user.graphql
var UpdateUser string

//go:embed validate_token.graphql
var ValidateToken string
