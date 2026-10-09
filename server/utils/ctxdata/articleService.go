package ctxdata

import "context"

type articleCtxKey string

const (
	UserID   = articleCtxKey("userID")
	UserRole = articleCtxKey("userRole")
)

func (c *CtxData) StoreUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserID, userID)
}

func (c *CtxData) GetUserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(UserID).(string)
	return v, ok
}
