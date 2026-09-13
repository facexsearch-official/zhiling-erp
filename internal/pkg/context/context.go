package context

import "context"

type contextKey string

const (
	keyTenantID contextKey = "tenant_id"
	keyUserID   contextKey = "user_id"
	keyShopID   contextKey = "shop_id"
	keyRole     contextKey = "role"
	keyIsOwner  contextKey = "is_owner"
)

func WithTenant(ctx context.Context, tenantID, userID, shopID int64, role int, isOwner bool) context.Context {
	ctx = context.WithValue(ctx, keyTenantID, tenantID)
	ctx = context.WithValue(ctx, keyUserID, userID)
	ctx = context.WithValue(ctx, keyShopID, shopID)
	ctx = context.WithValue(ctx, keyRole, role)
	ctx = context.WithValue(ctx, keyIsOwner, isOwner)
	return ctx
}

func GetTenantID(ctx context.Context) int64 {
	v, _ := ctx.Value(keyTenantID).(int64)
	return v
}

func GetUserID(ctx context.Context) int64 {
	v, _ := ctx.Value(keyUserID).(int64)
	return v
}

func GetShopID(ctx context.Context) int64 {
	v, _ := ctx.Value(keyShopID).(int64)
	return v
}

func GetRole(ctx context.Context) int {
	v, _ := ctx.Value(keyRole).(int)
	return v
}

func GetIsOwner(ctx context.Context) bool {
	v, _ := ctx.Value(keyIsOwner).(bool)
	return v
}
