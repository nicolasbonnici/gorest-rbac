package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

// Repository handles database operations for RBAC
type Repository struct {
	db database.Database
}

func NewRepository(db database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	qb := query.New(r.db.Dialect()).
		Select("r.name").
		From("user_roles").As("ur").
		JoinAs("roles", "r", query.ColEq("ur.role_id", "r.id")).
		Where(query.Eq("ur.user_id", userID)).
		OrderBy("r.name", query.ASC)

	queryStr, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return roles, nil
}

func (r *Repository) AssignRole(ctx context.Context, userID, roleName, assignedBy string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	if _, err := uuid.Parse(assignedBy); err != nil {
		return fmt.Errorf("invalid assignedBy ID format: %w", err)
	}

	// Find role ID
	selectQb := query.New(r.db.Dialect()).
		Select("id").
		From("roles").
		Where(query.Eq("name", roleName))

	selectQuery, selectArgs, err := selectQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build select query: %w", err)
	}

	var roleID string
	err = r.db.QueryRow(ctx, selectQuery, selectArgs...).Scan(&roleID)
	if err == sql.ErrNoRows {
		if logErr := r.logAudit(ctx, userID, "promote", roleName, assignedBy, false, "role not found"); logErr != nil {
			return fmt.Errorf("role not found (audit log failed: %v)", logErr)
		}
		return ErrRoleNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}

	// Insert user role assignment
	insertQb := query.New(r.db.Dialect()).
		Insert("user_roles").
		Columns("user_id", "role_id", "assigned_by", "assigned_at").
		Values(userID, roleID, assignedBy, time.Now())

	insertQuery, insertArgs, err := insertQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	// Add ON CONFLICT DO NOTHING for idempotent insert
	insertQuery += " ON CONFLICT (user_id, role_id) DO NOTHING"

	_, err = r.db.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		if logErr := r.logAudit(ctx, userID, "promote", roleName, assignedBy, false, err.Error()); logErr != nil {
			return fmt.Errorf("failed to assign role: %w (audit log failed: %v)", err, logErr)
		}
		return fmt.Errorf("failed to assign role: %w", err)
	}

	if err := r.logAudit(ctx, userID, "promote", roleName, assignedBy, true, ""); err != nil {
		return fmt.Errorf("role assigned but audit log failed: %w", err)
	}

	return nil
}

func (r *Repository) RemoveRole(ctx context.Context, userID, roleName, removedBy string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	if _, err := uuid.Parse(removedBy); err != nil {
		return fmt.Errorf("invalid removedBy ID format: %w", err)
	}

	// Find role ID
	selectQb := query.New(r.db.Dialect()).
		Select("id").
		From("roles").
		Where(query.Eq("name", roleName))

	selectQuery, selectArgs, err := selectQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build select query: %w", err)
	}

	var roleID string
	err = r.db.QueryRow(ctx, selectQuery, selectArgs...).Scan(&roleID)
	if err == sql.ErrNoRows {
		if logErr := r.logAudit(ctx, userID, "demote", roleName, removedBy, false, "role not found"); logErr != nil {
			return fmt.Errorf("role not found (audit log failed: %v)", logErr)
		}
		return ErrRoleNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}

	// Delete user role assignment
	deleteQb := query.New(r.db.Dialect()).
		Delete("user_roles").
		Where(query.And(
			query.Eq("user_id", userID),
			query.Eq("role_id", roleID),
		))

	deleteQuery, deleteArgs, err := deleteQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.Exec(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		if logErr := r.logAudit(ctx, userID, "demote", roleName, removedBy, false, err.Error()); logErr != nil {
			return fmt.Errorf("failed to remove role: %w (audit log failed: %v)", err, logErr)
		}
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		if logErr := r.logAudit(ctx, userID, "demote", roleName, removedBy, false, "user does not have role"); logErr != nil {
			return fmt.Errorf("user does not have role %s (audit log failed: %v)", roleName, logErr)
		}
		return fmt.Errorf("user does not have role %s", roleName)
	}

	if err := r.logAudit(ctx, userID, "demote", roleName, removedBy, true, ""); err != nil {
		return fmt.Errorf("role removed but audit log failed: %w", err)
	}

	return nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]Role, error) {
	qb := query.New(r.db.Dialect()).
		Select("id", "name", "description", "parent", "created_at").
		From("roles").
		OrderBy("name", query.ASC)

	queryStr, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var roles []Role
	for rows.Next() {
		var role Role
		var parent sql.NullString

		err := rows.Scan(&role.ID, &role.Name, &role.Description, &parent, &role.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		if parent.Valid {
			role.Parent = parent.String
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return roles, nil
}

func (r *Repository) GetRoleHierarchy(ctx context.Context) (map[string][]string, error) {
	qb := query.New(r.db.Dialect()).
		Select("parent_role", "child_role").
		From("role_hierarchy").
		OrderBy("parent_role", query.ASC).
		OrderBy("child_role", query.ASC)

	queryStr, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query role hierarchy: %w", err)
	}
	defer func() { _ = rows.Close() }()

	hierarchy := make(map[string][]string)
	for rows.Next() {
		var parent, child string
		if err := rows.Scan(&parent, &child); err != nil {
			return nil, fmt.Errorf("failed to scan hierarchy: %w", err)
		}

		hierarchy[parent] = append(hierarchy[parent], child)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return hierarchy, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]UserRoles, error) {
	// This query uses PostgreSQL-specific ARRAY_AGG, so we use RawExpr for the aggregate
	qb := query.New(r.db.Dialect()).
		Select("ur.user_id").
		SelectExpr(
			query.As(query.Max(query.Col("ur.assigned_at")), "updated_at"),
			query.RawExpr("ARRAY_AGG(r.name ORDER BY r.name) as roles"),
		).
		From("user_roles").As("ur").
		JoinAs("roles", "r", query.ColEq("ur.role_id", "r.id")).
		GroupBy("ur.user_id").
		OrderBy("ur.user_id", query.ASC)

	queryStr, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []UserRoles
	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var userRoles UserRoles
		var rolesArray []string

		if err := rows.Scan(&userRoles.UserID, &userRoles.UpdatedAt, &rolesArray); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		userRoles.Roles = rolesArray
		users = append(users, userRoles)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

func (r *Repository) CreateRole(ctx context.Context, name, description, parent string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert role - handle parent as NULL if empty
	var parentValue any
	if parent == "" {
		parentValue = nil
	} else {
		parentValue = parent
	}

	insertQb := query.New(r.db.Dialect()).
		Insert("roles").
		Columns("id", "name", "description", "parent").
		Values(uuid.New().String(), name, description, parentValue)

	insertQuery, insertArgs, err := insertQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = tx.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	// Create hierarchy entry if parent is specified
	if parent != "" {
		hierQb := query.New(r.db.Dialect()).
			Insert("role_hierarchy").
			Columns("parent_role", "child_role").
			Values(parent, name)

		hierQuery, hierArgs, err := hierQb.Build()
		if err != nil {
			return fmt.Errorf("failed to build hierarchy insert query: %w", err)
		}

		// Add ON CONFLICT DO NOTHING
		hierQuery += " ON CONFLICT DO NOTHING"

		_, err = tx.Exec(ctx, hierQuery, hierArgs...)
		if err != nil {
			return fmt.Errorf("failed to create hierarchy: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) DeleteRole(ctx context.Context, name string) error {
	deleteQb := query.New(r.db.Dialect()).
		Delete("roles").
		Where(query.Eq("name", name))

	deleteQuery, deleteArgs, err := deleteQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.Exec(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRoleNotFound
	}

	return nil
}

func (r *Repository) logAudit(ctx context.Context, userID, action, role, actor string, success bool, errorMsg string) error {
	// Handle empty error message as NULL
	var errorValue any
	if errorMsg == "" {
		errorValue = nil
	} else {
		errorValue = errorMsg
	}

	insertQb := query.New(r.db.Dialect()).
		Insert("rbac_audit_log").
		Columns("user_id", "action", "role", "actor", "success", "error").
		Values(userID, action, role, actor, success, errorValue)

	insertQuery, insertArgs, err := insertQb.Build()
	if err != nil {
		return fmt.Errorf("failed to build audit insert query: %w", err)
	}

	_, err = r.db.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("failed to log audit entry: %w", err)
	}

	return nil
}

func (r *Repository) GetAuditLog(ctx context.Context, limit int) ([]AuditEntry, error) {
	qb := query.New(r.db.Dialect()).
		Select("id", "timestamp", "user_id", "action", "role", "actor", "success").
		SelectExpr(query.As(query.Coalesce(query.Col("error"), query.Literal("")), "error")).
		From("rbac_audit_log").
		OrderBy("timestamp", query.DESC).
		Limit(limit)

	queryStr, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.UserID,
			&entry.Action,
			&entry.Role,
			&entry.Actor,
			&entry.Success,
			&entry.Error,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}
