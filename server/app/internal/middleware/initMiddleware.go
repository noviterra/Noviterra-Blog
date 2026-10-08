package middleware

import jwtauth "Noviterra/server/utils/jwtAuth"

type UserGormOP interface {
	GetUserAccountStatus(id string) (int, error)
	GetUserRole(id string) (int, error)
	CheckIDExists(id string) (bool, error)
}

type Middleware struct {
	authorization *jwtauth.Authentication
	userRepo      UserGormOP
}

func NewMiddleware(authorization *jwtauth.Authentication, userRepo UserGormOP) *Middleware {
	return &Middleware{
		authorization: authorization,
		userRepo:      userRepo,
	}
}
