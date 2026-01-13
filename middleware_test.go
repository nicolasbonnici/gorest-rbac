package rbac

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type mockRoleProvider struct {
	roles []string
	err   error
}

func (m *mockRoleProvider) GetRoles(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.roles, nil
}

func TestMiddleware(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	tests := []struct {
		name           string
		roleProvider   RoleProvider
		expectedStatus int
		expectedRoles  []string
	}{
		{
			name: "successful role extraction",
			roleProvider: &mockRoleProvider{
				roles: []string{"user", "editor"},
			},
			expectedStatus: 200,
			expectedRoles:  []string{"user", "editor"},
		},
		{
			name: "no roles",
			roleProvider: &mockRoleProvider{
				roles: []string{},
			},
			expectedStatus: 200,
			expectedRoles:  []string{},
		},
		{
			name: "role provider error",
			roleProvider: &mockRoleProvider{
				err: ErrPermissionDenied,
			},
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(Middleware(&voter, tt.roleProvider))

			app.Get("/test", func(c *fiber.Ctx) error {
				roles := c.Locals("user_roles")
				if roles == nil && len(tt.expectedRoles) > 0 {
					t.Error("expected roles in context, got nil")
				}
				return c.SendString("success")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	config.RoleHierarchy = map[string][]string{
		"admin":  {"editor"},
		"editor": {"user"},
	}

	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	tests := []struct {
		name           string
		userRoles      []string
		requiredRoles  []string
		expectedStatus int
		providerError  error
	}{
		{
			name:           "user has required role",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"admin"},
			expectedStatus: 200,
		},
		{
			name:           "user has one of required roles",
			userRoles:      []string{"editor"},
			requiredRoles:  []string{"admin", "editor"},
			expectedStatus: 200,
		},
		{
			name:           "user has inherited role",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"user"},
			expectedStatus: 200,
		},
		{
			name:           "user lacks required role",
			userRoles:      []string{"user"},
			requiredRoles:  []string{"admin"},
			expectedStatus: 403,
		},
		{
			name:           "superuser bypasses check",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"nonexistent"},
			expectedStatus: 200,
		},
		{
			name:           "role provider error",
			userRoles:      []string{},
			requiredRoles:  []string{"admin"},
			expectedStatus: 401,
			providerError:  ErrPermissionDenied,
		},
		{
			name:           "no user roles",
			userRoles:      []string{},
			requiredRoles:  []string{"admin"},
			expectedStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			roleProvider := &mockRoleProvider{
				roles: tt.userRoles,
				err:   tt.providerError,
			}

			app.Use(RequireRole(&voter, roleProvider, tt.requiredRoles...))

			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("success")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	type TestBody struct {
		PublicField string `json:"public" rbac:"read:*;write:*"`
		AdminField  string `json:"admin" rbac:"read:admin;write:admin"`
	}

	tests := []struct {
		name           string
		method         string
		userRoles      []string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET request - skip validation",
			method:         "GET",
			userRoles:      []string{},
			body:           `{"admin":"value"}`,
			expectedStatus: 200,
		},
		{
			name:           "POST with public field only",
			method:         "POST",
			userRoles:      []string{"user"},
			body:           `{"public":"value"}`,
			expectedStatus: 200,
		},
		{
			name:           "POST with admin field - user role",
			method:         "POST",
			userRoles:      []string{"user"},
			body:           `{"admin":"value"}`,
			expectedStatus: 403,
		},
		{
			name:           "POST with admin field - admin role",
			method:         "POST",
			userRoles:      []string{"admin"},
			body:           `{"admin":"value"}`,
			expectedStatus: 200,
		},
		{
			name:           "PUT with forbidden field",
			method:         "PUT",
			userRoles:      []string{"user"},
			body:           `{"admin":"value"}`,
			expectedStatus: 403,
		},
		{
			name:           "PATCH with allowed field",
			method:         "PATCH",
			userRoles:      []string{"user"},
			body:           `{"public":"value"}`,
			expectedStatus: 200,
		},
		{
			name:           "invalid JSON body",
			method:         "POST",
			userRoles:      []string{"admin"},
			body:           `{invalid}`,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			roleProvider := &mockRoleProvider{
				roles: tt.userRoles,
			}

			resourceType := &TestBody{}
			app.Use(ValidateRequest(&voter, roleProvider, resourceType))

			app.All("/test", func(c *fiber.Ctx) error {
				return c.SendString("success")
			})

			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestFilterResponse(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "GET request with roles",
			method:         "GET",
			userRoles:      []string{"user"},
			expectedStatus: 200,
		},
		{
			name:           "GET request without roles",
			method:         "GET",
			userRoles:      []string{},
			expectedStatus: 200,
		},
		{
			name:           "POST request - skip filtering",
			method:         "POST",
			userRoles:      []string{"user"},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			roleProvider := &mockRoleProvider{
				roles: tt.userRoles,
			}

			app.Use(FilterResponse(&voter, roleProvider))

			app.All("/test", func(c *fiber.Ctx) error {
				return c.SendString("success")
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	config.RoleHierarchy = map[string][]string{
		"admin": {"user"},
	}

	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	app := fiber.New()

	roleProvider := &mockRoleProvider{
		roles: []string{"user"},
	}

	app.Use(Middleware(&voter, roleProvider))

	adminOnly := RequireRole(&voter, roleProvider, "admin")
	userAllowed := RequireRole(&voter, roleProvider, "user")

	app.Get("/public", func(c *fiber.Ctx) error {
		return c.SendString("public")
	})

	app.Get("/user", userAllowed, func(c *fiber.Ctx) error {
		return c.SendString("user")
	})

	app.Get("/admin", adminOnly, func(c *fiber.Ctx) error {
		return c.SendString("admin")
	})

	tests := []struct {
		path           string
		expectedStatus int
	}{
		{"/public", 200},
		{"/user", 200},
		{"/admin", 403},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestMiddlewareContextPropagation(t *testing.T) {
	ClearCache()

	config := DefaultConfig()
	voter, err := NewVoter(config)
	if err != nil {
		t.Fatalf("failed to create voter: %v", err)
	}

	expectedRoles := []string{"user", "editor"}

	roleProvider := &mockRoleProvider{
		roles: expectedRoles,
	}

	app := fiber.New()
	app.Use(Middleware(&voter, roleProvider))

	app.Get("/test", func(c *fiber.Ctx) error {
		roles := c.Locals("user_roles")
		if roles == nil {
			t.Error("expected roles in locals, got nil")
			return c.SendStatus(500)
		}

		rolesSlice, ok := roles.([]string)
		if !ok {
			t.Errorf("expected []string, got %T", roles)
			return c.SendStatus(500)
		}

		if len(rolesSlice) != len(expectedRoles) {
			t.Errorf("expected %d roles, got %d", len(expectedRoles), len(rolesSlice))
			return c.SendStatus(500)
		}

		return c.SendString("success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
