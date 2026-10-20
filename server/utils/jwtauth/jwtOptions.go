package jwtauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func (a *Authentication) GenerateDoubleToken(id string) (string, string, error) {
	accessToken, err := a.GenerateAccessToken(id)
	if err != nil {
		return "", "", fmt.Errorf("AccessToken生成失败：%v", err.Error())
	}

	refreshToken, err := a.generateRefreshToken(id)
	if err != nil {
		return "", "", fmt.Errorf("RefreshToken生成失败：%v", err.Error())
	}

	return accessToken, refreshToken, nil
}

func (a *Authentication) GenerateAccessToken(id string) (string, error) {
	claim := UserClaims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(a.AccessExpire))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Noviterra_AccessToken",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(a.AccessSecret))
}

func (a *Authentication) generateRefreshToken(id string) (string, error) {
	claim := UserClaims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(a.RefreshExpire))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Noviterra_RefreshToken",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(a.RefreshSecret))
}

func (a *Authentication) ParseAccessToken(tokenStr string) (*UserClaims, error) {
	// 尝试解析AccessToken
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// 身份校验
	if claim, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claim, nil
	}
	return nil, errors.New("Token验证失败！")
}

func (a *Authentication) ParseRefreshToken(tokenStr string) (*UserClaims, error) {
	// 尝试解析AccessToken
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.RefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// 身份校验
	if claim, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claim, nil
	}
	return nil, errors.New("Token验证失败！")
}

func (a *Authentication) ReturnRefreshTokenExpire() int64 {
	return a.RefreshExpire
}
