# GoREST RBAC Plugin

[![CI](https://github.com/nicolasbonnici/gorest-rbac/workflows/CI/badge.svg)](https://github.com/nicolasbonnici/gorest-rbac/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-rbac)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-rbac)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, annotation-driven Role-Based Access Control (RBAC) plugin for GoREST. Secure your API endpoints with minimal boilerplate using simple struct tags.

## Features

- **Annotation-based permissions** - Define access control using struct tags
- **Field-level granularity** - Control read/write access per field
- **Role hierarchy** - Roles inherit permissions from child roles
- **Superuser support** - Admin role with unrestricted access
- **High performance** - Annotation caching and minimal overhead
- **GoREST native** - Seamless integration with GoREST framework
- **CLI tool** - Manage user roles from command line
- **Database-backed** - PostgreSQL and SQLite support

## Installation

```bash
go get github.com/nicolasbonnici/gorest-rbac
```


## Development Environment

To set up your development environment:

```bash
make install
```

This will:
- Install Go dependencies
- Install development tools (golangci-lint)
- Set up git hooks (pre-commit linting and tests)

## Quick Start

### 1. Define Resources with RBAC Tags

```go
type Article struct {
    ID      int    `json:"id" rbac:"read:*;write:none"`
    Title   string `json:"title" rbac:"read:*;write:editor,admin"`
    Content string `json:"content" rbac:"read:*;write:editor,admin"`
    Status  string `json:"status" rbac:"read:editor,admin;write:admin"`
}
```

### 2. Configure RBAC

```go
cfg := rbac.Config{
    DefaultPolicy: rbac.DenyAll,
    SuperuserRole: "admin",
    RoleHierarchy: map[string][]string{
        "admin":  {"editor"},
        "editor": {"user"},
    },
}

voter, _ := rbac.NewVoter(cfg)
roleProvider := rbac.NewFiberRoleProvider("user_roles", "user_id")
```

### 3. Add Middleware

```go
app := fiber.New()
app.Use(rbac.Middleware(voter, roleProvider))
```

### 4. Use in Handlers

```go
app.Get("/articles/:id", func(c *fiber.Ctx) error {
    ctx := c.UserContext()
    article := fetchArticle()

    filtered, _ := voter.FilterRead(ctx, &article)
    return c.JSON(filtered)
})
```

## Annotation Syntax

The `rbac` struct tag uses the format: `rbac:"read:roles;write:roles"`

### Special Keywords

- `*` - All roles (public access)
- `any` - Any authenticated user
- `none` - No access

### Examples

```go
type Resource struct {
    PublicField    string `rbac:"read:*;write:admin"`           // Public read, admin write
    UserField      string `rbac:"read:user,admin;write:admin"`  // User/admin read, admin write
    AdminOnlyField string `rbac:"read:admin;write:admin"`       // Admin only
    ReadOnlyField  string `rbac:"read:*;write:none"`            // Read-only
    HiddenField    string `rbac:"read:none;write:none"`         // Hidden (admin can still access)
}
```

## Role Hierarchy

Define role inheritance to reduce configuration:

```yaml
rbac:
  role_hierarchy:
    admin:
      - editor
      - moderator
    editor:
      - user
```

With this hierarchy:
- `admin` inherits `editor` and `moderator` permissions
- `editor` inherits `user` permissions
- A user with `admin` role effectively has `[admin, editor, moderator, user]`

## Superuser Admin Role

The admin role (configurable) bypasses all permission checks:

```go
cfg := rbac.Config{
    SuperuserRole: "admin",  // This role has unrestricted access
}
```

Even fields marked `rbac:"read:none;write:none"` are accessible to the superuser role.

## CLI Tool

Manage user roles using the `rbac-cli` command:

### Installation

```bash
go install github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli@latest
```

### Usage

```bash
rbac-cli users list
rbac-cli users show john@example.com
rbac-cli users promote john@example.com editor
rbac-cli users demote john@example.com editor
rbac-cli roles list
rbac-cli roles hierarchy
```

### Configuration

Create `.rbac-cli.yaml`:

```yaml
database:
  type: postgres
  host: localhost
  port: 5432
  name: myapp
  user: postgres
  password: ${DB_PASSWORD}
```

## GoREST Plugin Integration

### gorest.yaml

```yaml
plugins:
  - name: rbac
    enabled: true
    config:
      default_policy: deny_all
      superuser_role: admin
      role_hierarchy:
        admin: [editor, moderator]
        editor: [user]
      cache_enabled: true
      cache_ttl: 300
      strict_mode: true
```

### Registering the Plugin

```go
import (
    "github.com/nicolasbonnici/gorest/pluginloader"
    rbacplugin "github.com/nicolasbonnici/gorest-rbac"
)

func init() {
    pluginloader.RegisterPluginFactory("rbac", rbacplugin.NewPlugin)
}
```

## API Reference

### Voter Interface

```go
type Voter interface {
    CheckRead(ctx context.Context, resource interface{}, field string) error
    CheckWrite(ctx context.Context, resource interface{}, field string) error
    FilterRead(ctx context.Context, resource interface{}) (interface{}, error)
    ValidateWrite(ctx context.Context, resource interface{}) error
}
```

### Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DefaultPolicy` | Policy | `DenyAll` | Default authorization policy |
| `SuperuserRole` | string | `"admin"` | Role with unrestricted access |
| `RoleHierarchy` | map[string][]string | `{}` | Parent-child role relationships |
| `CacheEnabled` | bool | `true` | Enable annotation caching |
| `CacheTTL` | int | `300` | Cache TTL in seconds |
| `StrictMode` | bool | `true` | Fail on invalid annotations |
| `DefaultFieldPolicy` | string | `"deny"` | Policy for fields without tags |

## Database Schema

The plugin creates these tables:

- `roles` - Available roles
- `user_roles` - User-role assignments
- `role_hierarchy` - Role parent-child relationships
- `rbac_audit_log` - Audit trail of role changes

## Examples

See the [examples](./examples) directory:

- [Basic Example](./examples/basic) - Simple RBAC usage
- More examples coming soon

## Testing

```bash
make test
make test-coverage
make lint
```

## Building

```bash
make build
make build-cli
```

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md).

---

## Git Hooks

This directory contains git hooks for the GoREST plugin to maintain code quality.

### Available Hooks

#### pre-commit

Runs before each commit to ensure code quality:
- **Linting**: Runs `make lint` to check code style and potential issues
- **Tests**: Runs `make test` to verify all tests pass

### Installation

#### Automatic Installation

Run the install script from the project root:

```bash
./.githooks/install.sh
```

#### Manual Installation

Copy the hooks to your `.git/hooks` directory:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---


## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- [Documentation](./docs)
- [Issues](https://github.com/nicolasbonnici/gorest-rbac/issues)
- [Discussions](https://github.com/nicolasbonnici/gorest-rbac/discussions)
