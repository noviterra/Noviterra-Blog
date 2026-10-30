package ctxdata

import "context"

type ctxKeyID string
type ctxKeyRole string

const (
	ctxkeyUserID   ctxKeyID   = "user_id"
	ctxKeyUserRole ctxKeyRole = "user_role"
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
