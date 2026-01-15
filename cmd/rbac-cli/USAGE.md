# RBAC CLI - Complete Usage Guide

## Table of Contents

1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Command Reference](#command-reference)
4. [Examples](#examples)
5. [Output Formats](#output-formats)
6. [Database Setup](#database-setup)
7. [Troubleshooting](#troubleshooting)

## Installation

### Quick Install

```bash
cd cmd/rbac-cli
make build
```

### Install to System

```bash
cd cmd/rbac-cli
make install
```

The binary will be installed to `$GOPATH/bin/rbac-cli`.

## Configuration

### Configuration Precedence

The CLI uses the following order of precedence (highest to lowest):

1. Command-line flags (`--db-url`, `--output`)
2. Environment variables (`RBAC_DB_URL`, `RBAC_OUTPUT`)
3. Configuration file (`.rbac-cli.yaml`)

### Configuration File

Create a `.rbac-cli.yaml` file:

```yaml
database_url: postgres://user:password@localhost:5432/mydb?sslmode=disable
output: table
```

Or use the init command:

```bash
rbac-cli config init --db-url "postgres://localhost/mydb"
```

### Environment Variables

```bash
export RBAC_DB_URL="postgres://localhost:5432/mydb"
export RBAC_OUTPUT="json"
```

## Command Reference

### Global Flags

- `--config string`: Config file path (default: `.rbac-cli.yaml`)
- `--db-url string`: Database connection string
- `--output string` / `-o`: Output format (table/json/yaml)

### Users Commands

#### `rbac-cli users list`

List all users with their assigned roles.

**Flags:**
- All global flags

**Examples:**
```bash
rbac-cli users list
rbac-cli users list --output json
rbac-cli users list -o yaml
```

**Output (table):**
```
+----------+------------------+----------------------+
| User ID  | Roles            | Updated At           |
+----------+------------------+----------------------+
| user-123 | admin, moderator | 2024-01-12T10:30:00Z |
| user-456 | viewer           | 2024-01-12T09:15:00Z |
+----------+------------------+----------------------+
```

---

#### `rbac-cli users show <user-id>`

Show roles for a specific user.

**Arguments:**
- `user-id`: The user identifier

**Flags:**
- All global flags

**Examples:**
```bash
rbac-cli users show user-123
rbac-cli users show user-123 --output json
```

**Output (JSON):**
```json
{
  "user_id": "user-123",
  "roles": ["admin", "moderator"]
}
```

---

#### `rbac-cli users promote <user-id> <role>`

Assign a role to a user.

**Arguments:**
- `user-id`: The user identifier
- `role`: The role name to assign

**Flags:**
- `--actor string`: Actor performing the promotion (default: "cli-admin")
- All global flags

**Examples:**
```bash
rbac-cli users promote user-123 admin
rbac-cli users promote user-456 moderator --actor "superadmin"
```

**Output:**
```
Successfully assigned role 'admin' to user 'user-123'
```

---

#### `rbac-cli users demote <user-id> <role>`

Remove a role from a user.

**Arguments:**
- `user-id`: The user identifier
- `role`: The role name to remove

**Flags:**
- `--actor string`: Actor performing the demotion (default: "cli-admin")
- All global flags

**Examples:**
```bash
rbac-cli users demote user-123 moderator
rbac-cli users demote user-456 admin --actor "superadmin"
```

**Output:**
```
Successfully removed role 'moderator' from user 'user-123'
```

---

### Roles Commands

#### `rbac-cli roles list`

List all available roles in the system.

**Flags:**
- All global flags

**Examples:**
```bash
rbac-cli roles list
rbac-cli roles list --output json
rbac-cli roles list -o yaml
```

**Output (table):**
```
+-----------+---------------------------+--------+----------------------+
| Name      | Description               | Parent | Created At           |
+-----------+---------------------------+--------+----------------------+
| admin     | Administrator role        | (root) | 2024-01-01T00:00:00Z |
| moderator | Moderator role            | viewer | 2024-01-01T00:00:00Z |
| viewer    | Read-only viewer role     | (root) | 2024-01-01T00:00:00Z |
+-----------+---------------------------+--------+----------------------+
```

---

#### `rbac-cli roles hierarchy`

Display the role hierarchy as a tree.

**Flags:**
- All global flags

**Examples:**
```bash
rbac-cli roles hierarchy
rbac-cli roles hierarchy --output json
```

**Output (table/tree):**
```
Role Hierarchy:
==================================================
admin
├── moderator
│   └── viewer
└── editor
    └── contributor

superuser
```

**Output (JSON):**
```json
{
  "admin": ["moderator", "editor"],
  "moderator": ["viewer"],
  "editor": ["contributor"]
}
```

---

### Config Commands

#### `rbac-cli config init`

Initialize a new configuration file.

**Flags:**
- `--db-url string`: Database connection string
- `--output string`: Default output format (default: "table")
- `--force`: Overwrite existing config file

**Examples:**
```bash
rbac-cli config init
rbac-cli config init --db-url "postgres://localhost/mydb"
rbac-cli config init --force
```

**Output:**
```
Configuration file created: .rbac-cli.yaml

Please edit the file to set your database connection string.
```

---

#### `rbac-cli config show`

Show the current configuration.

**Examples:**
```bash
rbac-cli config show
```

**Output:**
```
Current Configuration:
==================================================
Config file:    /home/user/project/.rbac-cli.yaml
                (loaded)

Settings:
Database URL:   postgres://loc...***...
Output format:  table

Environment Variables:
RBAC_DB_URL:    (not set)
RBAC_OUTPUT:    (not set)
```

---

#### `rbac-cli config path`

Show the path to the configuration file.

**Examples:**
```bash
rbac-cli config path
```

**Output:**
```
/home/user/project/.rbac-cli.yaml
```

---

## Examples

### Daily Operations

#### Check User Permissions

```bash
# View all users and their roles
rbac-cli users list

# Check specific user's roles
rbac-cli users show user-789
```

#### Promote/Demote Users

```bash
# Promote user to admin
rbac-cli users promote user-123 admin --actor "hr-admin"

# Demote user from moderator
rbac-cli users demote user-456 moderator --actor "hr-admin"
```

#### Audit and Reporting

```bash
# Export all users to JSON for auditing
rbac-cli users list --output json > audit-$(date +%Y%m%d).json

# Export role hierarchy
rbac-cli roles hierarchy --output yaml > hierarchy.yaml

# Get list of all roles
rbac-cli roles list --output json > roles.json
```

### Automation Scripts

#### Bash: Bulk User Promotion

```bash
#!/bin/bash
# promote_users.sh - Promote multiple users to a role

ROLE="viewer"
USERS="user-001 user-002 user-003"

for user in $USERS; do
    echo "Promoting $user to $ROLE..."
    rbac-cli users promote "$user" "$ROLE" --actor "automation"
done
```

#### Python: Generate User Report

```python
#!/usr/bin/env python3
# user_report.py - Generate HTML report of users and roles

import subprocess
import json
from datetime import datetime

# Fetch users
result = subprocess.run(
    ['rbac-cli', 'users', 'list', '--output', 'json'],
    capture_output=True,
    text=True
)

users = json.loads(result.stdout)

# Generate HTML report
html = f"""
<html>
<head><title>RBAC User Report - {datetime.now()}</title></head>
<body>
<h1>User Roles Report</h1>
<table border="1">
<tr><th>User ID</th><th>Roles</th><th>Updated At</th></tr>
"""

for user in users:
    roles = ', '.join(user['roles'])
    html += f"""
<tr>
    <td>{user['user_id']}</td>
    <td>{roles}</td>
    <td>{user['updated_at']}</td>
</tr>
"""

html += """
</table>
</body>
</html>
"""

with open('user_report.html', 'w') as f:
    f.write(html)

print("Report generated: user_report.html")
```

### CI/CD Integration

#### GitLab CI Example

```yaml
# .gitlab-ci.yml
rbac-audit:
  stage: audit
  script:
    - rbac-cli users list --output json > users.json
    - rbac-cli roles list --output json > roles.json
  artifacts:
    paths:
      - users.json
      - roles.json
    expire_in: 30 days
```

#### GitHub Actions Example

```yaml
# .github/workflows/rbac-audit.yml
name: RBAC Audit

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - name: Install RBAC CLI
        run: |
          go install github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli@latest

      - name: Export Users
        env:
          RBAC_DB_URL: ${{ secrets.RBAC_DB_URL }}
        run: |
          rbac-cli users list --output json > users.json

      - name: Upload Artifact
        uses: actions/upload-artifact@v2
        with:
          name: rbac-audit
          path: users.json
```

## Output Formats

### Table Format

Best for: Terminal viewing, human readability

```bash
rbac-cli users list --output table
```

Features:
- Bordered tables
- Aligned columns
- Auto-wrapped text
- Color support (if terminal supports it)

### JSON Format

Best for: Automation, scripting, API integration

```bash
rbac-cli users list --output json
```

Features:
- Pretty-printed with 2-space indentation
- Valid JSON format
- Easy to parse with `jq` or programming languages

Example with `jq`:
```bash
rbac-cli users list --output json | jq '.[] | select(.roles | contains(["admin"]))'
```

### YAML Format

Best for: Configuration files, documentation

```bash
rbac-cli users list --output yaml
```

Features:
- Human-friendly format
- Easy to read and edit
- Compatible with YAML parsers

## Database Setup

### Required Tables

The CLI requires the following database tables:

- `roles`: Role definitions
- `user_roles`: User-role assignments
- `role_hierarchy`: Role parent-child relationships
- `rbac_audit_log`: Audit trail

### Schema

See the main RBAC plugin documentation for the complete database schema.

### PostgreSQL Setup

```sql
-- Run migrations from the main RBAC plugin
-- Tables will be created automatically
```

### SQLite Setup

```bash
# Create SQLite database
sqlite3 rbac.db < migrations/schema.sql
```

## Troubleshooting

### Database Connection Errors

**Problem:** `failed to connect to database`

**Solutions:**
1. Verify database URL format
2. Check database is running
3. Verify credentials
4. Check network connectivity
5. For PostgreSQL, try `?sslmode=disable`

**Example:**
```bash
# Test connection
rbac-cli --db-url "postgres://user:pass@localhost:5432/db?sslmode=disable" roles list
```

---

### Permission Errors

**Problem:** `permission denied` or `insufficient permissions`

**Solutions:**
1. Verify database user has SELECT/INSERT/DELETE permissions
2. Check RBAC tables exist
3. Ensure actor has rights to modify roles

---

### Config File Not Loading

**Problem:** Config file seems to be ignored

**Solutions:**
1. Check you're in the correct directory
2. Verify file name is exactly `.rbac-cli.yaml`
3. Check YAML syntax is valid
4. Use `--config` flag to specify full path

**Debug:**
```bash
# Check config path
rbac-cli config path

# Show loaded config
rbac-cli config show
```

---

### Role Not Found

**Problem:** `role not found` error

**Solutions:**
1. List all available roles: `rbac-cli roles list`
2. Check spelling and case sensitivity
3. Verify role exists in database

---

### Empty Results

**Problem:** Commands return no data

**Solutions:**
1. Verify database has data
2. Check you're connected to the correct database
3. Run migrations if tables are empty

**Check:**
```bash
# List all roles
rbac-cli roles list

# List all users
rbac-cli users list
```

---

## Advanced Usage

### Custom Config Location

```bash
rbac-cli --config /etc/rbac/config.yaml users list
```

### Multiple Databases

Use different config files for different environments:

```bash
# Production
rbac-cli --config .rbac-cli.prod.yaml users list

# Staging
rbac-cli --config .rbac-cli.staging.yaml users list

# Development
rbac-cli --config .rbac-cli.dev.yaml users list
```

### Scripting Best Practices

1. Always use `--output json` for parsing
2. Check exit codes: `$?` (bash) or `$LASTEXITCODE` (PowerShell)
3. Set environment variables for automation
4. Use `--actor` flag to track who made changes

Example:
```bash
#!/bin/bash
set -e  # Exit on error

export RBAC_DB_URL="postgres://localhost/mydb"
export RBAC_OUTPUT="json"

# This will exit if it fails
rbac-cli users promote user-123 admin

echo "Promotion successful"
```

---

## Performance Tips

1. Use JSON output for large datasets (faster than table rendering)
2. For bulk operations, connect once and reuse the connection
3. Consider database indexes on frequently queried columns
4. Use connection pooling for concurrent operations

---

## Security Considerations

1. Never commit `.rbac-cli.yaml` with real credentials
2. Use environment variables in CI/CD
3. Restrict database user permissions to minimum required
4. Always specify `--actor` for audit trail
5. Regularly review audit logs

---

## Support

For issues, questions, or contributions:

- GitHub Issues: https://github.com/nicolasbonnici/gorest-rbac
- Documentation: See main RBAC plugin README
