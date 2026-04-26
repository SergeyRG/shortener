package contextutils

import "context"

type ctxKey int

const (
	UserIDKey ctxKey = iota
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
