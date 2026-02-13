package rbac

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func Middleware(voter Voter, roleProvider RoleProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract roles from request
		ctx := c.UserContext()
		roles, err := roleProvider.GetRoles(ctx)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "failed to extract roles")
		}

		// Add roles to context for downstream handlers
		c.Locals("user_roles", roles)

		ctxWithValues := context.WithValue(ctx, fiberContextKey, c)
		ctxWithValues = WithRoles(ctxWithValues, roles)
		c.SetUserContext(ctxWithValues)

		// For write operations (POST, PUT, PATCH, DELETE), we'll validate in the handler
		// For now, just pass through
		return c.Next()
	}
}

func RequireRole(voter Voter, roleProvider RoleProvider, requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		roles, err := roleProvider.GetRoles(ctx)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "failed to extract roles")
		}

		ctxWithRoles := WithRoles(ctx, roles)
		c.SetUserContext(ctxWithRoles)

		if voter.IsSuperuser(roles) {
			return c.Next()
		}

		config := voter.GetConfig()
		if !HasAnyRole(roles, requiredRoles, config.RoleHierarchy) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}

		return c.Next()
	}
}

func ValidateRequest(voter Voter, roleProvider RoleProvider, resourceType interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only validate write operations
		method := c.Method()
		if method != fiber.MethodPost && method != fiber.MethodPut && method != fiber.MethodPatch {
			return c.Next()
		}

		// Extract roles
		ctx := c.UserContext()
		roles, err := roleProvider.GetRoles(ctx)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "failed to extract roles")
		}

		ctxWithRoles := WithRoles(ctx, roles)
		c.SetUserContext(ctxWithRoles)

		// Parse request body into resource type
		if err := c.BodyParser(resourceType); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		// Validate write permissions
		if err := voter.ValidateWrite(ctxWithRoles, resourceType); err != nil {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}

		return c.Next()
	}
}

func FilterResponse(voter Voter, roleProvider RoleProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only filter read operations
		if c.Method() != fiber.MethodGet {
			return c.Next()
		}

		// Extract roles
		ctx := c.UserContext()
		roles, err := roleProvider.GetRoles(ctx)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "failed to extract roles")
		}

		ctxWithRoles := WithRoles(ctx, roles)
		c.SetUserContext(ctxWithRoles)

		// Continue with the handler
		// Note: Actual response filtering would need to be done in the handler
		// or with a response interceptor
		return c.Next()
	}
}
