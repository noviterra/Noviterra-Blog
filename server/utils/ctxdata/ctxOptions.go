package ctxdata

import "context"

const (
	ctxkeyUserID   = "user_id"
	ctxKeyUserRole = "user_role"

	ctxUserLoginStatus = "user_login_status"
)

func (c *CtxData) StoreUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxkeyUserID, userID)
}

func (c *CtxData) StoreUserRoleInContext(ctx context.Context, userRole int) context.Context {
	return context.WithValue(ctx, ctxKeyUserRole, userRole)
}

func (c *CtxData) GetUserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxkeyUserID).(string); ok {
		return v
	}
	return ""
}

func (c *CtxData) GetUserRoleFromContext(ctx context.Context) int {
	if v, ok := ctx.Value(ctxKeyUserRole).(int); ok {
		return v
	}
	return -1
}

func (c *CtxData) SetUserLoginStatus(ctx context.Context, isLogin bool) context.Context {
	return context.WithValue(ctx, ctxUserLoginStatus, isLogin)
}

func (c *CtxData) GetUserLoginStatus(ctx context.Context) bool {
	if v, ok := ctx.Value(ctxUserLoginStatus).(bool); ok {
		return v
	}
	return false
}
