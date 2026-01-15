# RBAC CLI Tool

A comprehensive command-line interface for managing Role-Based Access Control (RBAC) in the GoREST framework.

## Features

- List all users with their assigned roles
- View specific user's roles
- Promote users by assigning roles
- Demote users by removing roles
- List all available roles in the system
- Display role hierarchy as a tree
- Configuration file support
- Multiple output formats (table, JSON, YAML)
- Support for PostgreSQL and SQLite databases

## Installation

### From Source

```bash
cd cmd/rbac-cli
go build -o rbac-cli .
```

### Install to $GOPATH/bin

```bash
cd cmd/rbac-cli
go install .
```

## Quick Start

1. Initialize a configuration file:

```bash
rbac-cli config init
```

2. Edit the `.rbac-cli.yaml` file with your database connection string:

```yaml
database_url: postgres://user:password@localhost:5432/mydb?sslmode=disable
output: table
```

3. Run commands:

```bash
# List all users
rbac-cli users list

# Show user roles
rbac-cli users show user-123

# Promote a user
rbac-cli users promote user-123 admin

# List all roles
rbac-cli roles list
```

## Configuration

The CLI supports multiple configuration methods (in order of precedence):

1. **Command-line flags**: `--db-url`, `--output`
2. **Environment variables**: `RBAC_DB_URL`, `RBAC_OUTPUT`
3. **Configuration file**: `.rbac-cli.yaml` (current directory or specified with `--config`)

### Configuration File

Create a `.rbac-cli.yaml` file:

```yaml
database_url: postgres://localhost:5432/mydb
output: table  # Options: table, json, yaml
```

Initialize with defaults:

```bash
rbac-cli config init
```

View current configuration:

```bash
rbac-cli config show
```

## Database Support

### PostgreSQL

```bash
# Using command-line flag
rbac-cli --db-url "postgres://user:password@localhost:5432/dbname?sslmode=disable" users list

# Using environment variable
export RBAC_DB_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
rbac-cli users list
```

### SQLite

```bash
# Using file path
rbac-cli --db-url "rbac.db" users list

# Using file:// protocol
rbac-cli --db-url "file:rbac.sqlite" users list
```

## Commands

### Users Management

#### List All Users

```bash
rbac-cli users list
rbac-cli users list --output json
rbac-cli users list --output yaml
```

Output example (table):
```
+----------+------------------+----------------------+
| User ID  | Roles            | Updated At           |
+----------+------------------+----------------------+
| user-123 | admin, moderator | 2024-01-12T10:30:00Z |
| user-456 | viewer           | 2024-01-12T09:15:00Z |
+----------+------------------+----------------------+
```

#### Show User Roles

```bash
rbac-cli users show user-123
rbac-cli users show user-123 --output json
```

#### Promote User (Assign Role)

```bash
rbac-cli users promote user-123 admin
rbac-cli users promote user-123 moderator --actor "admin-user"
```

#### Demote User (Remove Role)

```bash
rbac-cli users demote user-123 moderator
rbac-cli users demote user-123 admin --actor "superadmin"
```

### Roles Management

#### List All Roles

```bash
rbac-cli roles list
rbac-cli roles list --output json
```

Output example (table):
```
+-----------+---------------------------+--------+----------------------+
| Name      | Description               | Parent | Created At           |
+-----------+---------------------------+--------+----------------------+
| admin     | Administrator role        | (root) | 2024-01-01T00:00:00Z |
| moderator | Moderator role            | viewer | 2024-01-01T00:00:00Z |
| viewer    | Read-only viewer role     | (root) | 2024-01-01T00:00:00Z |
+-----------+---------------------------+--------+----------------------+
```

#### Display Role Hierarchy

```bash
rbac-cli roles hierarchy
```

Output example:
```
Role Hierarchy:
==================================================
admin
├── moderator
│   └── viewer
└── editor
    └── contributor
```

### Configuration Management

#### Initialize Config

```bash
rbac-cli config init
rbac-cli config init --db-url "postgres://localhost/mydb"
rbac-cli config init --force  # Overwrite existing config
```

#### Show Current Configuration

```bash
rbac-cli config show
```

#### Get Config File Path

```bash
rbac-cli config path
```

## Output Formats

The CLI supports three output formats:

### Table (Default)

Human-readable table format, ideal for terminal viewing.

```bash
rbac-cli users list
rbac-cli users list --output table
```

### JSON

Machine-readable JSON format, perfect for scripting and automation.

```bash
rbac-cli users list --output json
```

### YAML

Human-friendly YAML format, easy to read and parse.

```bash
rbac-cli users list --output yaml
```

## Examples

### Complete Workflow

```bash
# 1. Initialize configuration
rbac-cli config init --db-url "postgres://localhost/rbac_db"

# 2. List all roles
rbac-cli roles list

# 3. View role hierarchy
rbac-cli roles hierarchy

# 4. List all users
rbac-cli users list

# 5. Promote a user to admin
rbac-cli users promote user-123 admin --actor "superadmin"

# 6. Verify the change
rbac-cli users show user-123

# 7. Export users to JSON
rbac-cli users list --output json > users.json

# 8. Demote user
rbac-cli users demote user-123 admin --actor "superadmin"
```

### Scripting Examples

#### Bash: Export all users to JSON

```bash
#!/bin/bash
export RBAC_DB_URL="postgres://localhost/mydb"
rbac-cli users list --output json > users_$(date +%Y%m%d).json
```

#### Python: Process user data

```python
import subprocess
import json

result = subprocess.run(
    ['rbac-cli', 'users', 'list', '--output', 'json'],
    capture_output=True,
    text=True,
    env={'RBAC_DB_URL': 'postgres://localhost/mydb'}
)

users = json.loads(result.stdout)
for user in users:
    print(f"User {user['user_id']} has roles: {', '.join(user['roles'])}")
```

## Building

### Development Build

```bash
go build -o rbac-cli .
```

### Production Build with Version Info

```bash
VERSION="1.0.0"
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build \
  -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o rbac-cli .
```

### Cross-Platform Builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o rbac-cli-linux-amd64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -o rbac-cli-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -o rbac-cli-darwin-arm64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o rbac-cli-windows-amd64.exe .
```

## Troubleshooting

### Database Connection Issues

If you see "failed to connect to database":

1. Verify your database URL is correct
2. Check that the database is running and accessible
3. Ensure your user has proper permissions
4. For PostgreSQL, verify SSL settings (add `?sslmode=disable` if needed)

### Permission Denied

If you see permission errors:

1. Ensure the CLI has necessary database permissions
2. Check that the actor (if specified) has rights to modify roles
3. Verify database user has SELECT, INSERT, DELETE permissions on RBAC tables

### Config File Not Found

If the config file isn't loading:

1. Check you're in the correct directory
2. Use `--config` flag to specify the full path
3. Use `rbac-cli config path` to see where it's looking

## Environment Variables

- `RBAC_DB_URL`: Database connection string
- `RBAC_OUTPUT`: Default output format (table, json, yaml)

## Exit Codes

- `0`: Success
- `1`: Error occurred

## License

This CLI tool is part of the GoREST RBAC plugin.

## Support

For issues and questions, please refer to the main GoREST RBAC documentation.
