package model

type UserType string

const (
	UserTypeOIDC      UserType = "openid_connect"
	UserTypeAnonymous UserType = "anonymous"
)

type User struct {
	ID   string
	Type UserType
}
