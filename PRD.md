# Product Requirements Document: GoREST RBAC Plugin

## 1. Executive Summary

### 1.1 Product Overview
GoREST RBAC is a lightweight, annotation-driven Role-Based Access Control (RBAC) plugin specifically designed for the GoREST library. It provides declarative access control through Go struct tags, enabling developers to secure GoREST API endpoints with minimal boilerplate code.

### 1.2 Target Users
- Go backend developers using GoREST library
- Teams requiring fine-grained access control for GoREST-based microservices
- Organizations implementing multi-tenant applications with GoREST
- API developers building secure REST APIs with GoREST

### 1.3 Key Value Proposition
- **Zero boilerplate**: Secure endpoints using simple annotations
- **Context-aware**: Separate controls for read and write operations
- **GoREST native**: Seamless integration with GoREST library
- **Performance focused**: Minimal overhead with efficient permission checking
- **Easy management**: Powerful CLI tool for managing roles with promote/demote commands

---

## 2. Goals and Objectives

### 2.1 Primary Goals
1. Provide annotation-based RBAC for GoREST library
2. Support distinct permission checks for read and write operations
3. Enable seamless integration with GoREST resource handlers
4. Maintain high performance with negligible latency overhead
5. Provide easy-to-use CLI tool for role management operations

### 2.2 Success Metrics
- Integration time < 30 minutes for new projects
- Permission check latency < 1ms for typical use cases
- Support for 100+ concurrent permission checks
- 95% reduction in boilerplate authorization code
- CLI role operations complete in < 5 seconds for typical use cases

---

## 3. Functional Requirements

### 3.1 Core RBAC Functionality

#### 3.1.1 Role Definition
- **REQ-001**: Support for defining custom roles (e.g., admin, user, editor, viewer)
- **REQ-002**: Role hierarchy support (optional)
- **REQ-003**: Multiple roles per user
- **REQ-004**: Dynamic role loading from configuration or database

#### 3.1.2 Permission Model
- **REQ-005**: Read permission checks (GET operations)
- **REQ-006**: Write permission checks (POST, PUT, PATCH, DELETE operations)
- **REQ-007**: Custom permission definitions beyond read/write
- **REQ-008**: Resource-level permissions
- **REQ-009**: Field-level permissions (optional for future)

### 3.2 Annotation-Based Configuration

#### 3.2.1 Struct Tag Annotations
```go
type UserResource struct {
    ID        int    `rbac:"read:user,admin;write:admin"`
    Email     string `rbac:"read:user,admin;write:admin"`
    Password  string `rbac:"read:none;write:admin"`
    CreatedAt time.Time `rbac:"read:user,admin;write:none"`
}
```

- **REQ-010**: Support `rbac` struct tag for field-level permissions
- **REQ-011**: Syntax: `rbac:"read:role1,role2;write:role3,role4"`
- **REQ-012**: Special keywords: `none` (no access), `any` (all authenticated), `*` (all roles)
- **REQ-013**: Default behavior when tag is missing (configurable: allow-all or deny-all)

#### 3.2.2 Resource/Endpoint Annotations
```go
type UserResource struct {
    gorest.Resource `rbac:"read:user,admin;write:admin"`
    // ... fields
}
```

- **REQ-014**: Resource-level annotations on GoREST resource structs
- **REQ-015**: Automatic enforcement through GoREST middleware
- **REQ-016**: Override struct-level permissions with resource-level annotations

### 3.3 GoREST Integration

#### 3.3.1 GoREST Middleware
- **REQ-017**: Provide native middleware for GoREST library
- **REQ-018**: Automatic extraction of user context from GoREST request
- **REQ-019**: Support for custom user context providers
- **REQ-020**: Early request termination with 403 Forbidden on permission denial
- **REQ-021**: Integration with GoREST resource lifecycle hooks

#### 3.3.2 Context Management
- **REQ-022**: Extract user roles from `context.Context`
- **REQ-023**: Standard context key for user/role information
- **REQ-024**: Support for JWT token parsing (optional integration)
- **REQ-025**: Custom authenticator interface

### 3.4 Configuration

#### 3.4.1 Configuration Sources
- **REQ-026**: YAML configuration file support
- **REQ-027**: JSON configuration file support
- **REQ-028**: Environment variable configuration
- **REQ-029**: Programmatic configuration via Go code

#### 3.4.2 Configuration Options
```yaml
rbac:
  default_policy: deny_all  # or allow_all
  superuser_role: admin  # role with unrestricted access (default: "admin")
  role_hierarchy:
    admin: [editor, user]
    editor: [user]
  cache:
    enabled: true
    ttl: 300s
  strict_mode: true  # fail on invalid annotations
```

- **REQ-030**: Default policy configuration (allow-all vs deny-all)
- **REQ-031**: Role hierarchy definition
- **REQ-032**: Permission caching configuration
- **REQ-033**: Strict mode for validation
- **REQ-034**: Custom error messages per role/resource

### 3.5 Runtime Enforcement

#### 3.5.1 Permission Checking
- **REQ-035**: `CheckRead(user, resource, field)` method
- **REQ-036**: `CheckWrite(user, resource, field)` method
- **REQ-037**: `CheckCustom(user, resource, permission)` method
- **REQ-038**: Batch permission checking for multiple fields
- **REQ-039**: Explain mode: return why permission was denied

#### 3.5.2 Response Filtering
- **REQ-040**: Automatic field filtering in JSON responses
- **REQ-041**: Redaction of forbidden fields (null or omit)
- **REQ-042**: Support for nested struct filtering
- **REQ-043**: Array/slice element filtering

#### 3.5.3 Request Validation
- **REQ-044**: Validate incoming write requests against permissions
- **REQ-045**: Reject requests with forbidden fields before processing
- **REQ-046**: Detailed error messages indicating forbidden fields

#### 3.5.4 Superuser Role
- **REQ-047**: Support for "admin" superuser role with unrestricted access
- **REQ-048**: Admin role bypasses all permission checks (read and write)
- **REQ-049**: Configurable superuser role name (defaults to "admin")

### 3.6 CLI Tool for Role Management

#### 3.6.1 CLI Commands
- **REQ-070**: Command-line tool `rbac-cli` for managing user roles
- **REQ-071**: List all users and their roles
- **REQ-072**: List all available roles in the system
- **REQ-073**: Assign role to user (promote)
- **REQ-074**: Remove role from user (demote)
- **REQ-075**: Add new roles to the system
- **REQ-076**: Remove roles from the system
- **REQ-077**: Display role hierarchy

#### 3.6.2 CLI Command Examples
```bash
# List all users with their roles
rbac-cli users list

# List all available roles
rbac-cli roles list

# Show role hierarchy
rbac-cli roles hierarchy

# Assign role to user (promote)
rbac-cli users promote <user-id> <role>
rbac-cli users promote john@example.com editor

# Remove role from user (demote)
rbac-cli users demote <user-id> <role>
rbac-cli users demote john@example.com editor

# Add multiple roles at once
rbac-cli users promote <user-id> <role1,role2,role3>

# Show user's current roles
rbac-cli users show <user-id>

# Add new role to system
rbac-cli roles add <role-name> [--description "Role description"]

# Remove role from system
rbac-cli roles remove <role-name>

# Set role hierarchy
rbac-cli roles set-parent <child-role> <parent-role>
```

#### 3.6.3 CLI Configuration
- **REQ-078**: Support for configuration file (`.rbac-cli.yaml`)
- **REQ-079**: Database connection configuration for user/role storage
- **REQ-080**: Support for multiple backends (PostgreSQL, MySQL, MongoDB, SQLite)
- **REQ-081**: Environment variable support for sensitive configs
- **REQ-082**: Connection to remote RBAC service via API

#### 3.6.4 CLI Features
- **REQ-083**: Batch operations from CSV/JSON file
- **REQ-084**: Dry-run mode to preview changes
- **REQ-085**: Audit logging of all role changes
- **REQ-086**: Confirmation prompts for destructive operations
- **REQ-087**: JSON/YAML/Table output formats
- **REQ-088**: Filter and search capabilities
- **REQ-089**: Interactive mode for bulk operations

#### 3.6.5 CLI Output Examples
```bash
$ rbac-cli users list
┌────────────────────┬─────────────────────────┬──────────────────┐
│ USER ID            │ EMAIL                   │ ROLES            │
├────────────────────┼─────────────────────────┼──────────────────┤
│ 1                  │ admin@example.com       │ admin            │
│ 2                  │ john@example.com        │ editor, user     │
│ 3                  │ jane@example.com        │ user             │
└────────────────────┴─────────────────────────┴──────────────────┘

$ rbac-cli roles hierarchy
admin
  ├─ editor
  │   └─ user
  └─ moderator
      └─ user

$ rbac-cli users promote john@example.com admin --dry-run
[DRY RUN] Would add role 'admin' to user 'john@example.com'
Current roles: editor, user
New roles: admin, editor, user (admin inherits editor, user)

$ rbac-cli users promote john@example.com admin
✓ Successfully promoted john@example.com to admin
  Roles: admin, editor, user
```

#### 3.6.6 Batch Operations
- **REQ-090**: Import users and roles from CSV
- **REQ-091**: Export current role assignments to CSV/JSON
- **REQ-092**: Bulk promote/demote from file
- **REQ-093**: Rollback capability for batch operations

```bash
# Bulk import from CSV
rbac-cli users import roles.csv

# Export current assignments
rbac-cli users export --format json > roles-backup.json

# Bulk promote
rbac-cli users promote --from-file promotions.csv
```

---

## 4. Non-Functional Requirements

### 4.1 Performance
- **REQ-050**: Permission check latency < 1ms (p95)
- **REQ-051**: Support for in-memory permission cache
- **REQ-052**: Minimal memory footprint (< 10MB for typical usage)
- **REQ-053**: Zero allocations in hot path (after initialization)

### 4.2 Reliability
- **REQ-054**: Fail-safe: deny access on errors (configurable)
- **REQ-055**: Panic recovery in middleware
- **REQ-056**: Comprehensive error logging
- **REQ-057**: Graceful handling of malformed annotations

### 4.3 Security
- **REQ-058**: No permission bypass vulnerabilities
- **REQ-059**: Constant-time role comparison
- **REQ-060**: Input validation on all configuration
- **REQ-061**: Protection against role injection attacks

### 4.4 Usability
- **REQ-062**: Clear documentation with examples
- **REQ-063**: IDE autocomplete support for annotations
- **REQ-064**: Helpful error messages for misconfigurations
- **REQ-065**: GoREST integration guide

### 4.5 Maintainability
- **REQ-066**: Clean, idiomatic Go code
- **REQ-067**: Comprehensive unit test coverage (>80%)
- **REQ-068**: Integration test examples with GoREST
- **REQ-069**: Versioned API with semver

---

## 5. API Design

### 5.1 Core Interfaces

```go
// Authorizer is the main interface for RBAC checks
type Authorizer interface {
    CheckRead(ctx context.Context, resource interface{}, field string) error
    CheckWrite(ctx context.Context, resource interface{}, field string) error
    FilterRead(ctx context.Context, resource interface{}) (interface{}, error)
    ValidateWrite(ctx context.Context, resource interface{}) error
}

// RoleProvider extracts roles from request context
type RoleProvider interface {
    GetRoles(ctx context.Context) ([]string, error)
}

// Config holds RBAC configuration
type Config struct {
    DefaultPolicy   Policy
    SuperuserRole   string                 // Role with unrestricted access (default: "admin")
    RoleHierarchy   map[string][]string
    CacheEnabled    bool
    CacheTTL        time.Duration
    StrictMode      bool
}
```

### 5.2 GoREST Middleware Functions

```go
// GoREST middleware for RBAC enforcement
func Middleware(authorizer Authorizer) gorest.Middleware

// Resource-level interceptor
func ResourceInterceptor(authorizer Authorizer) gorest.Interceptor
```

### 5.3 Helper Functions

```go
// Create authorizer from config
func New(config Config) (Authorizer, error)

// Parse struct annotations
func ParseAnnotations(resource interface{}) (map[string]Permissions, error)

// Add to context
func WithRoles(ctx context.Context, roles []string) context.Context

// Get from context
func GetRoles(ctx context.Context) ([]string, bool)
```

### 5.4 CLI Role Management Interfaces

```go
// RoleManager handles role assignments and management
type RoleManager interface {
    // User role operations
    ListUsers(ctx context.Context, filter UserFilter) ([]UserRoles, error)
    GetUserRoles(ctx context.Context, userID string) ([]string, error)
    PromoteUser(ctx context.Context, userID string, roles []string) error
    DemoteUser(ctx context.Context, userID string, roles []string) error

    // Role operations
    ListRoles(ctx context.Context) ([]Role, error)
    GetRoleHierarchy(ctx context.Context) (*RoleTree, error)
    AddRole(ctx context.Context, role Role) error
    RemoveRole(ctx context.Context, roleName string) error
    SetRoleParent(ctx context.Context, childRole, parentRole string) error

    // Batch operations
    ImportRoles(ctx context.Context, data io.Reader, format Format) error
    ExportRoles(ctx context.Context, format Format) (io.Reader, error)

    // Audit
    GetAuditLog(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)
}

// UserRoles represents a user with their assigned roles
type UserRoles struct {
    UserID    string
    Email     string
    Roles     []string
    UpdatedAt time.Time
}

// Role represents a system role
type Role struct {
    Name        string
    Description string
    Parent      string
    CreatedAt   time.Time
}

// AuditEntry represents a role change audit log
type AuditEntry struct {
    Timestamp time.Time
    UserID    string
    Action    string // "promote", "demote", "add_role", "remove_role"
    Role      string
    Actor     string
    Success   bool
}
```

---

## 6. User Experience

### 6.1 Integration Flow

1. **Installation**
   ```bash
   go get github.com/username/gorest-rbac
   ```

2. **Define Resources with Annotations**
   ```go
   type Article struct {
       ID      int    `json:"id" rbac:"read:*;write:none"`
       Title   string `json:"title" rbac:"read:*;write:editor,admin"`
       Content string `json:"content" rbac:"read:*;write:editor,admin"`
       Status  string `json:"status" rbac:"read:editor,admin;write:admin"`
   }
   ```

3. **Configure RBAC**
   ```go
   cfg := rbac.Config{
       DefaultPolicy: rbac.DenyAll,
       RoleHierarchy: map[string][]string{
           "admin": {"editor"},
       },
   }
   auth := rbac.New(cfg)
   ```

4. **Add Middleware to GoREST**
   ```go
   service := gorest.NewService()
   service.Use(rbac.Middleware(auth))
   ```

5. **Define GoREST Resource**
   ```go
   type ArticleResource struct {
       gorest.Resource
       ID      int    `json:"id" rbac:"read:*;write:none"`
       Title   string `json:"title" rbac:"read:*;write:editor,admin"`
       Content string `json:"content" rbac:"read:*;write:editor,admin"`
       Status  string `json:"status" rbac:"read:editor,admin;write:admin"`
   }

   func (r *ArticleResource) Get(ctx gorest.Context) error {
       article := fetchArticle()
       filtered, _ := auth.FilterRead(ctx, article)
       return ctx.JSON(http.StatusOK, filtered)
   }
   ```

### 6.2 Common Use Cases

#### 6.2.1 Public Read, Authenticated Write
```go
type BlogPost struct {
    Content string `rbac:"read:any;write:authenticated"`
}
```

#### 6.2.2 Admin-Only Fields
```go
type User struct {
    Email    string `rbac:"read:self,admin;write:admin"`
    Password string `rbac:"read:none;write:self,admin"`
}
```

#### 6.2.3 Read-Only Computed Fields
```go
type Order struct {
    Total float64 `rbac:"read:user,admin;write:none"`
}
```

#### 6.2.4 Admin Superuser Access
```go
// Admin role bypasses all permission checks
type SensitiveData struct {
    Secret string `rbac:"read:none;write:none"`  // Admin can still access
    SSN    string `rbac:"read:compliance;write:none"`  // Admin can still access
}
```

### 6.3 CLI Tool Usage

#### 6.3.1 Initial Setup
```bash
# Install CLI tool
go install github.com/username/gorest-rbac/cmd/rbac-cli@latest

# Initialize configuration
rbac-cli init

# Configure database connection
rbac-cli config set-db postgres://user:pass@localhost/rbac

# Or use config file
cat > .rbac-cli.yaml <<EOF
database:
  type: postgres
  host: localhost
  port: 5432
  name: rbac
  user: admin
  password: secret
EOF
```

#### 6.3.2 Day-to-Day Operations
```bash
# Check user roles
rbac-cli users show john@example.com

# Promote user to editor
rbac-cli users promote john@example.com editor
✓ Successfully promoted john@example.com to editor

# Demote user
rbac-cli users demote john@example.com editor
✓ Successfully demoted john@example.com from editor

# Bulk operations from CSV
cat > promotions.csv <<EOF
user_id,role,action
john@example.com,editor,promote
jane@example.com,admin,promote
bob@example.com,moderator,demote
EOF

rbac-cli users batch --from-file promotions.csv
✓ Processed 3 role changes
  - 2 promotions
  - 1 demotion
```

#### 6.3.3 Role Management
```bash
# Add new role
rbac-cli roles add reviewer --description "Can review content"

# Set role hierarchy
rbac-cli roles set-parent reviewer editor

# View hierarchy
rbac-cli roles hierarchy
admin
  ├─ editor
  │   ├─ reviewer
  │   └─ user
  └─ moderator
```

---

## 7. Technical Architecture

### 7.1 Components

#### 7.1.1 Runtime Components (GoREST Integration)
```
┌─────────────────────────────────────────┐
│         HTTP Request                    │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      RBAC Middleware                    │
│  - Extract user/roles from context      │
│  - Check endpoint-level permissions     │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Handler Function                   │
│  - Business logic                       │
│  - Call FilterRead/ValidateWrite        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Authorizer                         │
│  - Parse struct annotations             │
│  - Check field-level permissions        │
│  - Filter/validate data                 │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Permission Cache (optional)        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         HTTP Response                   │
└─────────────────────────────────────────┘
```

#### 7.1.2 CLI Components
```
┌─────────────────────────────────────────┐
│         rbac-cli Command                │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      CLI Framework (cobra/cli)          │
│  - Parse arguments                      │
│  - Load configuration                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      RoleManager                        │
│  - User operations                      │
│  - Role operations                      │
│  - Batch operations                     │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Storage Backend                    │
│  - PostgreSQL / MySQL / MongoDB         │
│  - SQLite (local dev)                   │
│  - Remote API (optional)                │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Output Formatter                   │
│  - Table / JSON / YAML                  │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Console Output                  │
└─────────────────────────────────────────┘
```

### 7.2 Annotation Parser
- Reflection-based struct tag parsing
- Compile-time code generation option (future)
- Cache parsed annotations per type

### 7.3 Permission Evaluator
- Role hierarchy resolution
- Permission matching algorithm
- Fast path for common cases

---

## 8. Deliverables

### 8.1 Phase 1 (MVP)
- Core RBAC engine with read/write permissions
- Struct tag annotation parsing
- GoREST middleware and interceptors
- YAML configuration support
- Superuser admin role support
- Basic CLI tool with core commands:
  - `rbac-cli users list`
  - `rbac-cli users promote`
  - `rbac-cli users demote`
  - `rbac-cli roles list`
- Basic documentation and examples

### 8.2 Phase 2
- Permission caching
- Role hierarchy support
- Enhanced error messages
- Comprehensive test suite with GoREST integration tests
- Enhanced CLI features:
  - Batch operations (CSV/JSON import/export)
  - Dry-run mode
  - Multiple output formats (table, JSON, YAML)
  - Role hierarchy management commands
  - Audit logging
- Performance benchmarks

### 8.3 Phase 3
- Code generation for performance
- Field-level filtering for nested structs
- Custom permission types
- Advanced CLI features:
  - Interactive mode
  - Remote API support
  - Advanced filtering and search
  - Rollback capabilities
  - Integration with external auth providers
- Admin UI for permission management
- Real-time audit log streaming

---

## 9. Dependencies

### 9.1 Required Dependencies
- Go 1.21+
- GoREST library
- Standard library for core functionality

### 9.2 CLI Tool Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/olekukonko/tablewriter` - Table formatting
- Database drivers:
  - `github.com/lib/pq` - PostgreSQL
  - `github.com/go-sql-driver/mysql` - MySQL
  - `go.mongodb.org/mongo-driver` - MongoDB
  - `modernc.org/sqlite` - SQLite (pure Go)
- `gopkg.in/yaml.v3` - YAML config parsing
- `github.com/gocarina/gocsv` - CSV parsing

### 9.3 Optional Dependencies
- JWT libraries for token parsing (if JWT integration needed)
- `github.com/charmbracelet/bubbletea` - Interactive CLI mode (Phase 3)

---

## 10. Testing Strategy

### 10.1 Unit Tests
- Annotation parser tests
- Permission evaluator tests
- Middleware logic tests
- Configuration validation tests
- CLI command parsing tests
- RoleManager interface tests

### 10.2 Integration Tests
- End-to-end request flow tests with GoREST
- GoREST resource integration tests
- Performance benchmark tests
- CLI integration tests with test databases
- Batch operation tests (CSV/JSON import/export)

### 10.3 CLI-Specific Tests
- Command execution tests
- Database backend tests (PostgreSQL, MySQL, SQLite)
- Dry-run validation tests
- Output format tests (table, JSON, YAML)
- Error handling and rollback tests
- Audit log verification tests

### 10.4 Security Tests
- Permission bypass attempts
- Role injection testing
- Edge case scenarios
- CLI authentication and authorization tests

---

## 11. Documentation Requirements

### 11.1 User Documentation
- Quick start guide
- API reference
- Configuration guide
- GoREST integration guide
- CLI tool user guide:
  - Installation and setup
  - Command reference
  - Configuration options
  - Batch operations guide
  - Troubleshooting
- Best practices
- FAQ

### 11.2 CLI Documentation
- CLI command reference (man pages)
- Configuration file schema
- Database setup guides for each backend
- Batch operation file formats (CSV/JSON schemas)
- Migration guides from other RBAC systems
- Automation and scripting examples

### 11.3 Developer Documentation
- Architecture overview
- Contributing guide
- Code organization
- Testing guide
- CLI plugin development guide

---

## 12. Success Criteria

### 12.1 Launch Criteria
- [ ] Core RBAC functionality complete
- [ ] GoREST middleware and interceptors implemented
- [ ] Superuser admin role implemented
- [ ] CLI tool with core commands (list, promote, demote) working
- [ ] CLI supports at least PostgreSQL and SQLite backends
- [ ] 80%+ test coverage (including CLI tests)
- [ ] Documentation complete (library + CLI)
- [ ] Performance benchmarks meet targets
- [ ] Security review passed

### 12.2 Post-Launch Metrics
- GitHub stars and adoption rate
- Community contributions
- Issue resolution time
- Performance in production environments
- CLI tool download/installation count
- CLI command usage analytics (opt-in)
- Number of role changes managed via CLI

---

## 13. Open Questions

1. Should we support attribute-based access control (ABAC) in future versions?
2. Should role definitions be stored in code or external store?
3. Should we provide a UI for managing roles and permissions?
4. Integration with popular auth providers (Auth0, Keycloak, etc.)?
5. Should the admin superuser role be able to be disabled for certain resources?
6. Should the CLI support direct API access without database connection?
7. Should we provide pre-built binaries for multiple platforms?
8. Should the CLI have a daemon mode for real-time monitoring?
9. What's the upgrade path when role schema changes?
10. Should we support LDAP/Active Directory integration for role sync?

---

## 14. Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Performance overhead | High | Medium | Implement caching, benchmarking |
| Security vulnerabilities | High | Low | Security audit, extensive testing |
| Complex configuration | Medium | Medium | Sensible defaults, good docs |
| GoREST version compatibility | Medium | Low | Test with multiple GoREST versions |
| Adoption challenges | Medium | Medium | Great documentation, examples |
| CLI database compatibility issues | Medium | Medium | Support multiple backends, extensive testing |
| Accidental role changes via CLI | High | Medium | Dry-run mode, confirmations, audit logs |
| CLI dependency bloat | Low | Medium | Keep CLI tool separate, optional install |
| Data migration challenges | Medium | Low | Provide migration tools and docs |

---

## 15. Timeline Estimate

Note: This is a rough estimate based on complexity, not a commitment.

- Phase 1 MVP: Core functionality, GoREST middleware, admin role, basic CLI tool
- Phase 2 Enhancements: Caching, role hierarchy, enhanced error handling, advanced CLI features
- Phase 3 Advanced: Code generation, advanced features, interactive CLI, web UI

---

## Appendix A: Example Scenarios

### Scenario 1: Blog Platform
```go
type Post struct {
    ID        int       `json:"id" rbac:"read:*;write:none"`
    Title     string    `json:"title" rbac:"read:*;write:author,editor"`
    Content   string    `json:"content" rbac:"read:*;write:author,editor"`
    Published bool      `json:"published" rbac:"read:*;write:editor"`
    AuthorID  int       `json:"author_id" rbac:"read:*;write:none"`
}
```

### Scenario 2: Multi-tenant SaaS
```go
type TenantSettings struct {
    APIKey    string `json:"api_key" rbac:"read:tenant_admin;write:tenant_admin"`
    Plan      string `json:"plan" rbac:"read:tenant_admin;write:platform_admin"`
    Usage     int    `json:"usage" rbac:"read:tenant_admin,platform_admin;write:none"`
}
```

### Scenario 3: Healthcare Records
```go
type MedicalRecord struct {
    PatientID   string `json:"patient_id" rbac:"read:doctor,nurse;write:doctor"`
    Diagnosis   string `json:"diagnosis" rbac:"read:doctor;write:doctor"`
    Medications string `json:"medications" rbac:"read:doctor,nurse;write:doctor"`
    Notes       string `json:"notes" rbac:"read:doctor;write:doctor"`
}
```

---

## Appendix B: CLI Usage Examples

### B.1 Common Workflows

#### B.1.1 Onboarding New User
```bash
# Check if user exists
rbac-cli users show alice@company.com

# Promote to basic user role
rbac-cli users promote alice@company.com user
✓ Successfully promoted alice@company.com to user

# Verify
rbac-cli users show alice@company.com
User: alice@company.com
Roles: user
Last Updated: 2026-01-12 10:30:00
```

#### B.1.2 Promoting User to Editor
```bash
# Check current roles
rbac-cli users show bob@company.com
User: bob@company.com
Roles: user

# Promote with dry-run first
rbac-cli users promote bob@company.com editor --dry-run
[DRY RUN] Would add role 'editor' to user 'bob@company.com'
Current roles: user
New roles: editor, user

# Execute promotion
rbac-cli users promote bob@company.com editor
✓ Successfully promoted bob@company.com to editor
  Roles: editor, user
```

#### B.1.3 Bulk Promotions from CSV
```bash
# Create promotions file
cat > promotions.csv <<EOF
email,role,action
alice@company.com,editor,promote
bob@company.com,admin,promote
charlie@company.com,moderator,promote
EOF

# Preview changes
rbac-cli users batch --from-file promotions.csv --dry-run
[DRY RUN] Batch operation preview:
  - alice@company.com: promote to editor
  - bob@company.com: promote to admin
  - charlie@company.com: promote to moderator

Total: 3 operations

# Execute batch operation
rbac-cli users batch --from-file promotions.csv
✓ Processed 3 role changes
  - 3 promotions
  - 0 demotions
  - 0 errors

# View audit log
rbac-cli audit list --limit 10
┌─────────────────────┬───────────────────────┬─────────┬────────┬─────────┐
│ TIMESTAMP           │ USER                  │ ACTION  │ ROLE   │ ACTOR   │
├─────────────────────┼───────────────────────┼─────────┼────────┼─────────┤
│ 2026-01-12 10:35:00 │ charlie@company.com   │ promote │ mod... │ admin   │
│ 2026-01-12 10:35:00 │ bob@company.com       │ promote │ admin  │ admin   │
│ 2026-01-12 10:35:00 │ alice@company.com     │ promote │ editor │ admin   │
└─────────────────────┴───────────────────────┴─────────┴────────┴─────────┘
```

#### B.1.4 Managing Role Hierarchy
```bash
# View current hierarchy
rbac-cli roles hierarchy
admin
  └─ user

# Add new roles
rbac-cli roles add editor --description "Can edit content"
rbac-cli roles add moderator --description "Can moderate users"

# Set hierarchy
rbac-cli roles set-parent editor admin
rbac-cli roles set-parent moderator admin

# Verify
rbac-cli roles hierarchy
admin
  ├─ editor
  │   └─ user
  └─ moderator
      └─ user
```

#### B.1.5 Exporting Role Assignments
```bash
# Export to JSON
rbac-cli users export --format json > roles-backup.json

# Export to CSV
rbac-cli users export --format csv > roles-backup.csv

# Export filtered users
rbac-cli users export --role admin --format json > admin-users.json
```

### B.2 Database Configuration Examples

#### B.2.1 PostgreSQL
```yaml
# .rbac-cli.yaml
database:
  type: postgres
  host: localhost
  port: 5432
  name: rbac_db
  user: rbac_admin
  password: ${RBAC_DB_PASSWORD}  # from env var
  ssl_mode: require
```

#### B.2.2 MySQL
```yaml
# .rbac-cli.yaml
database:
  type: mysql
  host: localhost
  port: 3306
  name: rbac_db
  user: rbac_admin
  password: ${RBAC_DB_PASSWORD}
```

#### B.2.3 SQLite (Development)
```yaml
# .rbac-cli.yaml
database:
  type: sqlite
  path: ./rbac.db
```

### B.3 CSV File Formats

#### B.3.1 Batch Promotions/Demotions
```csv
email,role,action
user1@example.com,editor,promote
user2@example.com,admin,promote
user3@example.com,moderator,demote
```

#### B.3.2 Initial User Import
```csv
email,roles
admin@company.com,"admin,user"
editor1@company.com,"editor,user"
editor2@company.com,"editor,user"
user1@company.com,user
user2@company.com,user
```

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-12 | Initial | Initial PRD creation |
| 1.1 | 2026-01-12 | Updated | Focused on GoREST library only, added superuser admin role support |
| 1.2 | 2026-01-12 | Updated | Added comprehensive CLI tool for role management with promote/demote commands, batch operations, multi-database support, and audit logging |
