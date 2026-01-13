package rbac

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type RoleProvider interface {
	GetRoles(ctx context.Context) ([]string, error)
}

type DefaultRoleProvider struct{}

func NewDefaultRoleProvider() RoleProvider {
	return &DefaultRoleProvider{}
}

func (p *DefaultRoleProvider) GetRoles(ctx context.Context) ([]string, error) {
	roles, ok := GetRoles(ctx)
	if !ok {
		return []string{}, nil
	}
	return roles, nil
}

type FiberRoleProvider struct {
	RolesKey  string
	UserIDKey string
}

func NewFiberRoleProvider(rolesKey, userIDKey string) RoleProvider {
	if rolesKey == "" {
		rolesKey = "user_roles"
	}
	if userIDKey == "" {
		userIDKey = "user_id"
	}
	return &FiberRoleProvider{
		RolesKey:  rolesKey,
		UserIDKey: userIDKey,
	}
}

func (p *FiberRoleProvider) GetRoles(ctx context.Context) ([]string, error) {
	if fctx, ok := ctx.Value("fiber_ctx").(*fiber.Ctx); ok {
		if roles := fctx.Locals(p.RolesKey); roles != nil {
			switch v := roles.(type) {
			case []string:
				return v, nil
			case string:
				return []string{v}, nil
			}
		}
	}

	roles, ok := GetRoles(ctx)
	if !ok {
		return []string{}, nil
	}
	return roles, nil
}

type CustomRoleProvider struct {
	ExtractFunc func(context.Context) ([]string, error)
}

func NewCustomRoleProvider(extractFunc func(context.Context) ([]string, error)) RoleProvider {
	return &CustomRoleProvider{ExtractFunc: extractFunc}
}

func (p *CustomRoleProvider) GetRoles(ctx context.Context) ([]string, error) {
	if p.ExtractFunc == nil {
		return []string{}, nil
	}
	return p.ExtractFunc(ctx)
}
