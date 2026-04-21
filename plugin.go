package rbac

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
)

// RBACPlugin implements the GoREST plugin interface
type RBACPlugin struct {
	config       Config
	db           database.Database
	voter        Voter
	roleProvider RoleProvider
}

func NewPlugin() plugin.Plugin {
	return &RBACPlugin{}
}

func (p *RBACPlugin) Name() string {
	return "rbac"
}

func (p *RBACPlugin) Initialize(config map[string]interface{}) error {
	// Start with default config
	p.config = DefaultConfig()

	// Extract database if provided
	if db, ok := config["database"].(database.Database); ok {
		p.db = db
	}

	// Extract config values
	if defaultPolicy, ok := config["default_policy"].(string); ok {
		p.config.DefaultPolicy = Policy(defaultPolicy)
	}

	if superuserRole, ok := config["superuser_role"].(string); ok {
		p.config.SuperuserRole = superuserRole
	}

	if roleHierarchy, ok := config["role_hierarchy"].(map[string][]string); ok {
		p.config.RoleHierarchy = roleHierarchy
	} else if roleHierarchy, ok := config["role_hierarchy"].(map[string]interface{}); ok {
		converted, err := convertRoleHierarchy(roleHierarchy)
		if err != nil {
			return fmt.Errorf("invalid role_hierarchy configuration: %w", err)
		}
		p.config.RoleHierarchy = converted
	}

	if cacheEnabled, ok := config["cache_enabled"].(bool); ok {
		p.config.CacheEnabled = cacheEnabled
	}

	if cacheTTL, ok := config["cache_ttl"].(int); ok {
		p.config.CacheTTL = cacheTTL
	}

	if strictMode, ok := config["strict_mode"].(bool); ok {
		p.config.StrictMode = strictMode
	}

	if defaultFieldPolicy, ok := config["default_field_policy"].(string); ok {
		p.config.DefaultFieldPolicy = defaultFieldPolicy
	}

	// Validate configuration
	if err := p.config.Validate(); err != nil {
		return fmt.Errorf("invalid RBAC configuration: %w", err)
	}

	// Create voter
	voter, err := NewVoter(p.config)
	if err != nil {
		return fmt.Errorf("failed to create voter: %w", err)
	}
	p.voter = voter

	// Create default role provider
	p.roleProvider = NewFiberRoleProvider("user_roles", "user_id")

	return nil
}

func (p *RBACPlugin) Handler() fiber.Handler {
	return Middleware(p.voter, p.roleProvider)
}

func (p *RBACPlugin) SetupEndpoints(router fiber.Router) error {
	// Optionally setup management endpoints for roles
	// This could include endpoints to:
	// - List roles
	// - Assign/remove roles
	// - View role hierarchy
	// For now, we'll skip this as role management is primarily via CLI
	return nil
}

func (p *RBACPlugin) MigrationSource() interface{} {
	return GetMigrations()
}

func (p *RBACPlugin) MigrationDependencies() []string {
	// RBAC doesn't depend on other plugins
	return []string{}
}

func (p *RBACPlugin) GetVoter() Voter {
	return p.voter
}

func (p *RBACPlugin) GetConfig() Config {
	return p.config
}

func convertRoleHierarchy(input map[string]interface{}) (map[string][]string, error) {
	result := make(map[string][]string)

	for key, value := range input {
		switch v := value.(type) {
		case []string:
			result[key] = v
		case []interface{}:
			strSlice := make([]string, 0, len(v))
			for i, item := range v {
				str, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf("role_hierarchy[%s][%d]: expected string, got %T", key, i, item)
				}
				strSlice = append(strSlice, str)
			}
			result[key] = strSlice
		default:
			return nil, fmt.Errorf("role_hierarchy[%s]: expected []string or []interface{}, got %T", key, value)
		}
	}

	return result, nil
}
