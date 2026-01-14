package rbac

import (
	"context"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestNewDefaultRoleProvider(t *testing.T) {
	provider := NewDefaultRoleProvider()

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	_, ok := provider.(*DefaultRoleProvider)
	if !ok {
		t.Fatalf("expected *DefaultRoleProvider, got %T", provider)
	}
}

func TestDefaultRoleProviderGetRoles(t *testing.T) {
	provider := NewDefaultRoleProvider()

	tests := []struct {
		name          string
		ctx           context.Context
		expectedRoles []string
		expectError   bool
	}{
		{
			name:          "roles in context",
			ctx:           WithRoles(context.Background(), []string{"admin", "user"}),
			expectedRoles: []string{"admin", "user"},
			expectError:   false,
		},
		{
			name:          "no roles in context",
			ctx:           context.Background(),
			expectedRoles: []string{},
			expectError:   false,
		},
		{
			name:          "empty roles",
			ctx:           WithRoles(context.Background(), []string{}),
			expectedRoles: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := provider.GetRoles(tt.ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(roles, tt.expectedRoles) {
				t.Errorf("expected roles %v, got %v", tt.expectedRoles, roles)
			}
		})
	}
}

func TestNewFiberRoleProvider(t *testing.T) {
	tests := []struct {
		name             string
		rolesKey         string
		userIDKey        string
		expectedRolesKey string
		expectedUserIDKey string
	}{
		{
			name:             "default keys",
			rolesKey:         "",
			userIDKey:        "",
			expectedRolesKey: "user_roles",
			expectedUserIDKey: "user_id",
		},
		{
			name:             "custom keys",
			rolesKey:         "custom_roles",
			userIDKey:        "custom_user_id",
			expectedRolesKey: "custom_roles",
			expectedUserIDKey: "custom_user_id",
		},
		{
			name:             "partial custom keys",
			rolesKey:         "custom_roles",
			userIDKey:        "",
			expectedRolesKey: "custom_roles",
			expectedUserIDKey: "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewFiberRoleProvider(tt.rolesKey, tt.userIDKey)

			if provider == nil {
				t.Fatal("expected non-nil provider")
			}

			fiberProvider, ok := provider.(*FiberRoleProvider)
			if !ok {
				t.Fatalf("expected *FiberRoleProvider, got %T", provider)
			}

			if fiberProvider.RolesKey != tt.expectedRolesKey {
				t.Errorf("expected RolesKey %s, got %s", tt.expectedRolesKey, fiberProvider.RolesKey)
			}

			if fiberProvider.UserIDKey != tt.expectedUserIDKey {
				t.Errorf("expected UserIDKey %s, got %s", tt.expectedUserIDKey, fiberProvider.UserIDKey)
			}
		})
	}
}

func TestFiberRoleProviderGetRoles(t *testing.T) {
	provider := NewFiberRoleProvider("user_roles", "user_id")

	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedRoles []string
		expectError   bool
	}{
		{
			name: "roles from fiber context - string slice",
			setupContext: func() context.Context {
				app := fiber.New()
				reqCtx := &fasthttp.RequestCtx{}
				c := app.AcquireCtx(reqCtx)
				c.Locals("user_roles", []string{"admin", "user"})
				return context.WithValue(context.Background(), "fiber_ctx", c)
			},
			expectedRoles: []string{"admin", "user"},
			expectError:   false,
		},
		{
			name: "roles from fiber context - single string",
			setupContext: func() context.Context {
				app := fiber.New()
				reqCtx := &fasthttp.RequestCtx{}
				c := app.AcquireCtx(reqCtx)
				c.Locals("user_roles", "admin")
				return context.WithValue(context.Background(), "fiber_ctx", c)
			},
			expectedRoles: []string{"admin"},
			expectError:   false,
		},
		{
			name: "fallback to standard context",
			setupContext: func() context.Context {
				return WithRoles(context.Background(), []string{"user"})
			},
			expectedRoles: []string{"user"},
			expectError:   false,
		},
		{
			name: "no roles anywhere",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedRoles: []string{},
			expectError:   false,
		},
		{
			name: "fiber context with no roles",
			setupContext: func() context.Context {
				app := fiber.New()
				reqCtx := &fasthttp.RequestCtx{}
				c := app.AcquireCtx(reqCtx)
				return context.WithValue(context.Background(), "fiber_ctx", c)
			},
			expectedRoles: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			roles, err := provider.GetRoles(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(roles, tt.expectedRoles) {
				t.Errorf("expected roles %v, got %v", tt.expectedRoles, roles)
			}
		})
	}
}

func TestNewCustomRoleProvider(t *testing.T) {
	extractFunc := func(ctx context.Context) ([]string, error) {
		return []string{"custom"}, nil
	}

	provider := NewCustomRoleProvider(extractFunc)

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	customProvider, ok := provider.(*CustomRoleProvider)
	if !ok {
		t.Fatalf("expected *CustomRoleProvider, got %T", provider)
	}

	if customProvider.ExtractFunc == nil {
		t.Error("expected ExtractFunc to be set")
	}
}

func TestCustomRoleProviderGetRoles(t *testing.T) {
	tests := []struct {
		name          string
		extractFunc   func(context.Context) ([]string, error)
		expectedRoles []string
		expectError   bool
	}{
		{
			name: "successful extraction",
			extractFunc: func(ctx context.Context) ([]string, error) {
				return []string{"custom1", "custom2"}, nil
			},
			expectedRoles: []string{"custom1", "custom2"},
			expectError:   false,
		},
		{
			name: "extraction error",
			extractFunc: func(ctx context.Context) ([]string, error) {
				return nil, ErrPermissionDenied
			},
			expectedRoles: nil,
			expectError:   true,
		},
		{
			name:          "nil extract function",
			extractFunc:   nil,
			expectedRoles: []string{},
			expectError:   false,
		},
		{
			name: "empty roles",
			extractFunc: func(ctx context.Context) ([]string, error) {
				return []string{}, nil
			},
			expectedRoles: []string{},
			expectError:   false,
		},
		{
			name: "extract from context value",
			extractFunc: func(ctx context.Context) ([]string, error) {
				if roles, ok := ctx.Value("custom_roles").([]string); ok {
					return roles, nil
				}
				return []string{}, nil
			},
			expectedRoles: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCustomRoleProvider(tt.extractFunc)
			ctx := context.Background()

			roles, err := provider.GetRoles(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(roles, tt.expectedRoles) {
				t.Errorf("expected roles %v, got %v", tt.expectedRoles, roles)
			}
		})
	}
}

func TestCustomRoleProviderWithContextValue(t *testing.T) {
	extractFunc := func(ctx context.Context) ([]string, error) {
		if roles, ok := ctx.Value("custom_roles").([]string); ok {
			return roles, nil
		}
		return []string{"default"}, nil
	}

	provider := NewCustomRoleProvider(extractFunc)

	ctx := context.WithValue(context.Background(), "custom_roles", []string{"admin", "user"})

	roles, err := provider.GetRoles(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"admin", "user"}
	if !reflect.DeepEqual(roles, expected) {
		t.Errorf("expected roles %v, got %v", expected, roles)
	}

	emptyCtx := context.Background()
	roles, err = provider.GetRoles(emptyCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected = []string{"default"}
	if !reflect.DeepEqual(roles, expected) {
		t.Errorf("expected roles %v, got %v", expected, roles)
	}
}

func TestRoleProviderInterface(t *testing.T) {
	var _ RoleProvider = &DefaultRoleProvider{}
	var _ RoleProvider = &FiberRoleProvider{}
	var _ RoleProvider = &CustomRoleProvider{}
}
