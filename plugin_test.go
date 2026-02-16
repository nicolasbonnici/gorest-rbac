package rbac

import (
	"testing"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()

	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}

	rbacPlugin, ok := plugin.(*RBACPlugin)
	if !ok {
		t.Fatalf("expected *RBACPlugin, got %T", plugin)
	}

	if rbacPlugin.Name() != "rbac" {
		t.Errorf("expected name 'rbac', got %s", rbacPlugin.Name())
	}
}

func TestPluginName(t *testing.T) {
	plugin := &RBACPlugin{}
	name := plugin.Name()

	if name != "rbac" {
		t.Errorf("expected name 'rbac', got %s", name)
	}
}

func validateDefaultConfig(t *testing.T, p *RBACPlugin) {
	if p.config.DefaultPolicy != DenyAll {
		t.Errorf("expected default policy DenyAll, got %s", p.config.DefaultPolicy)
	}
	if p.config.SuperuserRole != "admin" {
		t.Errorf("expected superuser role 'admin', got %s", p.config.SuperuserRole)
	}
}

func validateCustomPolicy(t *testing.T, p *RBACPlugin) {
	if p.config.DefaultPolicy != AllowAll {
		t.Errorf("expected default policy AllowAll, got %s", p.config.DefaultPolicy)
	}
}

func validateSuperuserRole(t *testing.T, p *RBACPlugin) {
	if p.config.SuperuserRole != "superadmin" {
		t.Errorf("expected superuser role 'superadmin', got %s", p.config.SuperuserRole)
	}
}

func validateRoleHierarchy(t *testing.T, p *RBACPlugin) {
	if len(p.config.RoleHierarchy) != 2 {
		t.Errorf("expected 2 hierarchy entries, got %d", len(p.config.RoleHierarchy))
	}
	adminRoles, ok := p.config.RoleHierarchy["admin"]
	if !ok {
		t.Error("expected 'admin' in hierarchy")
		return
	}
	if len(adminRoles) != 2 {
		t.Errorf("expected 2 child roles for admin, got %d", len(adminRoles))
	}
}

func validateCacheSettings(t *testing.T, p *RBACPlugin) {
	if p.config.CacheEnabled {
		t.Error("expected cache to be disabled")
	}
	if p.config.CacheTTL != 600 {
		t.Errorf("expected cache TTL 600, got %d", p.config.CacheTTL)
	}
}

func validateStrictMode(t *testing.T, p *RBACPlugin) {
	if p.config.StrictMode {
		t.Error("expected strict mode to be disabled")
	}
}

func validateFieldPolicy(t *testing.T, p *RBACPlugin) {
	if p.config.DefaultFieldPolicy != "allow" {
		t.Errorf("expected default field policy 'allow', got %s", p.config.DefaultFieldPolicy)
	}
}

func validateCombinedSettings(t *testing.T, p *RBACPlugin) {
	if p.config.DefaultPolicy != AllowAll {
		t.Error("DefaultPolicy mismatch")
	}
	if p.config.SuperuserRole != "superadmin" {
		t.Error("SuperuserRole mismatch")
	}
	if !p.config.CacheEnabled {
		t.Error("CacheEnabled mismatch")
	}
	if p.config.CacheTTL != 600 {
		t.Error("CacheTTL mismatch")
	}
	if p.config.StrictMode {
		t.Error("StrictMode mismatch")
	}
	if p.config.DefaultFieldPolicy != "allow" {
		t.Error("DefaultFieldPolicy mismatch")
	}
	if len(p.config.RoleHierarchy) != 2 {
		t.Error("RoleHierarchy mismatch")
	}
}

func TestPluginInitialize(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		validate    func(*testing.T, *RBACPlugin)
	}{
		{
			name:        "empty config uses defaults",
			config:      map[string]interface{}{},
			expectError: false,
			validate:    validateDefaultConfig,
		},
		{
			name: "custom default policy",
			config: map[string]interface{}{
				"default_policy": "allow_all",
			},
			expectError: false,
			validate:    validateCustomPolicy,
		},
		{
			name: "custom superuser role",
			config: map[string]interface{}{
				"superuser_role": "superadmin",
			},
			expectError: false,
			validate:    validateSuperuserRole,
		},
		{
			name: "role hierarchy - map[string][]string",
			config: map[string]interface{}{
				"role_hierarchy": map[string][]string{
					"admin":  {"editor", "user"},
					"editor": {"user"},
				},
			},
			expectError: false,
			validate:    validateRoleHierarchy,
		},
		{
			name: "role hierarchy - map[string]interface{}",
			config: map[string]interface{}{
				"role_hierarchy": map[string]interface{}{
					"admin":  []interface{}{"editor", "user"},
					"editor": []interface{}{"user"},
				},
			},
			expectError: false,
			validate:    validateRoleHierarchy,
		},
		{
			name: "cache settings",
			config: map[string]interface{}{
				"cache_enabled": false,
				"cache_ttl":     600,
			},
			expectError: false,
			validate:    validateCacheSettings,
		},
		{
			name: "strict mode",
			config: map[string]interface{}{
				"strict_mode": false,
			},
			expectError: false,
			validate:    validateStrictMode,
		},
		{
			name: "default field policy",
			config: map[string]interface{}{
				"default_field_policy": "allow",
			},
			expectError: false,
			validate:    validateFieldPolicy,
		},
		{
			name: "invalid config - empty superuser role",
			config: map[string]interface{}{
				"superuser_role": "",
			},
			expectError: true,
		},
		{
			name: "invalid config - circular hierarchy",
			config: map[string]interface{}{
				"role_hierarchy": map[string][]string{
					"admin":  {"editor"},
					"editor": {"admin"},
				},
			},
			expectError: true,
		},
		{
			name: "all settings combined",
			config: map[string]interface{}{
				"default_policy":       "allow_all",
				"superuser_role":       "superadmin",
				"cache_enabled":        true,
				"cache_ttl":            600,
				"strict_mode":          false,
				"default_field_policy": "allow",
				"role_hierarchy": map[string][]string{
					"superadmin": {"admin"},
					"admin":      {"user"},
				},
			},
			expectError: false,
			validate:    validateCombinedSettings,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runPluginInitTest(t, tt.config, tt.expectError, tt.validate)
		})
	}
}

func runPluginInitTest(t *testing.T, config map[string]interface{}, expectError bool, validate func(*testing.T, *RBACPlugin)) {
	plugin := &RBACPlugin{}
	err := plugin.Initialize(config)

	if expectError {
		if err == nil {
			t.Error("expected error, got nil")
		}
		return
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if validate != nil {
		validate(t, plugin)
	}

	if plugin.voter == nil {
		t.Error("expected voter to be initialized")
	}

	if plugin.roleProvider == nil {
		t.Error("expected role provider to be initialized")
	}
}

func TestConvertRoleHierarchy(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string][]string
	}{
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string][]string{},
		},
		{
			name: "already string slice",
			input: map[string]interface{}{
				"admin": []string{"user", "editor"},
			},
			expected: map[string][]string{
				"admin": {"user", "editor"},
			},
		},
		{
			name: "interface slice",
			input: map[string]interface{}{
				"admin": []interface{}{"user", "editor"},
			},
			expected: map[string][]string{
				"admin": {"user", "editor"},
			},
		},
		{
			name: "mixed types",
			input: map[string]interface{}{
				"admin":  []string{"user"},
				"editor": []interface{}{"author", "contributor"},
			},
			expected: map[string][]string{
				"admin":  {"user"},
				"editor": {"author", "contributor"},
			},
		},
		{
			name: "single role",
			input: map[string]interface{}{
				"admin": []interface{}{"user"},
			},
			expected: map[string][]string{
				"admin": {"user"},
			},
		},
		{
			name: "empty slice",
			input: map[string]interface{}{
				"admin": []interface{}{},
			},
			expected: map[string][]string{
				"admin": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertRoleHierarchy(tt.input)
			if err != nil {
				t.Fatalf("convertRoleHierarchy failed: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}

			for key, expectedRoles := range tt.expected {
				actualRoles, ok := result[key]
				if !ok {
					t.Errorf("expected key %s not found in result", key)
					continue
				}

				if len(actualRoles) != len(expectedRoles) {
					t.Errorf("key %s: expected %d roles, got %d", key, len(expectedRoles), len(actualRoles))
					continue
				}

				for i, expectedRole := range expectedRoles {
					if actualRoles[i] != expectedRole {
						t.Errorf("key %s, index %d: expected %s, got %s", key, i, expectedRole, actualRoles[i])
					}
				}
			}
		})
	}
}

func TestConvertRoleHierarchyErrors(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "interface slice with non-string values",
			input: map[string]interface{}{
				"admin": []interface{}{"user", 123, "editor"},
			},
		},
		{
			name: "invalid type",
			input: map[string]interface{}{
				"admin": 123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertRoleHierarchy(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestPluginGetVoter(t *testing.T) {
	plugin := &RBACPlugin{}
	err := plugin.Initialize(map[string]interface{}{})
	if err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	voter := plugin.GetVoter()
	if voter == nil {
		t.Error("expected non-nil voter")
	}
}

func TestPluginGetConfig(t *testing.T) {
	plugin := &RBACPlugin{}

	config := map[string]interface{}{
		"default_policy": "allow_all",
		"superuser_role": "superadmin",
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	cfg := plugin.GetConfig()

	if cfg.DefaultPolicy != AllowAll {
		t.Errorf("expected AllowAll policy, got %s", cfg.DefaultPolicy)
	}

	if cfg.SuperuserRole != "superadmin" {
		t.Errorf("expected superuser role 'superadmin', got %s", cfg.SuperuserRole)
	}
}

func TestPluginMigrationDependencies(t *testing.T) {
	plugin := &RBACPlugin{}

	deps := plugin.MigrationDependencies()

	if deps == nil {
		t.Error("expected non-nil dependencies")
	}

	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(deps))
	}
}

func TestPluginHandler(t *testing.T) {
	plugin := &RBACPlugin{}
	err := plugin.Initialize(map[string]interface{}{})
	if err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	handler := plugin.Handler()
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}
