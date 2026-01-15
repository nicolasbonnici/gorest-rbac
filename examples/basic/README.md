# Basic RBAC Example

This example demonstrates how to use the GoREST RBAC plugin to protect API endpoints with role-based access control.

## Features

- Field-level permissions using struct tags
- Role hierarchy (admin inherits editor permissions)
- Automatic response filtering based on user roles
- Request validation for write operations

## Running the Example

```bash
cd examples/basic
go run main.go
```

The server will start on `http://localhost:3000`.

## Testing the Example

### As a regular user (read-only access to public fields)

```bash
curl -X GET http://localhost:3000/articles/1 \
  -H "X-User-Roles: user"
```

Response will include: ID, Title, Content, Author (but not Status)

### As an editor (can read all, write to some fields)

```bash
curl -X GET http://localhost:3000/articles/1 \
  -H "X-User-Roles: editor"
```

Response will include all fields including Status.

### As an admin (full access)

```bash
curl -X POST http://localhost:3000/articles \
  -H "X-User-Roles: admin" \
  -H "Content-Type: application/json" \
  -d '{"title":"New Article","content":"Content here","status":"draft"}'
```

Admin can write to all fields including Status.

## Permission Matrix

| Field   | User (Read) | Editor (Read) | Admin (Read) | User (Write) | Editor (Write) | Admin (Write) |
|---------|-------------|---------------|--------------|--------------|----------------|---------------|
| ID      | ✓           | ✓             | ✓            | ✗            | ✗              | ✗             |
| Title   | ✓           | ✓             | ✓            | ✗            | ✓              | ✓             |
| Content | ✓           | ✓             | ✓            | ✗            | ✓              | ✓             |
| Status  | ✗           | ✓             | ✓            | ✗            | ✗              | ✓             |
| Author  | ✓           | ✓             | ✓            | ✗            | ✗              | ✗             |
