package jwtauth

import (
	"github.com/golang-jwt/jwt/v4"
)

type UserClaims struct {
	ID string
	jwt.RegisteredClaims
}

type AuthenticationConfig struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshSecret string
	RefreshExpire int64
}

type Authentication struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshSecret string
	RefreshExpire int64
}

func NewAuthentication(authenticationConfig AuthenticationConfig) *Authentication {
	return &Authentication{
		AccessSecret:  authenticationConfig.AccessSecret,
		AccessExpire:  authenticationConfig.AccessExpire,
		RefreshSecret: authenticationConfig.RefreshSecret,
		RefreshExpire: authenticationConfig.RefreshExpire,
	}
}
