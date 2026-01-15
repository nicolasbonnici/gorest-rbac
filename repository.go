package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
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

	query := `
		SELECT r.name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := r.db.Query(ctx, query, userID)
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

	var roleID string
	err := r.db.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
	if err == sql.ErrNoRows {
		if logErr := r.logAudit(ctx, userID, "promote", roleName, assignedBy, false, "role not found"); logErr != nil {
			return fmt.Errorf("role not found (audit log failed: %v)", logErr)
		}
		return ErrRoleNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}

	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err = r.db.Exec(ctx, query, userID, roleID, assignedBy, time.Now())
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

	var roleID string
	err := r.db.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
	if err == sql.ErrNoRows {
		if logErr := r.logAudit(ctx, userID, "demote", roleName, removedBy, false, "role not found"); logErr != nil {
			return fmt.Errorf("role not found (audit log failed: %v)", logErr)
		}
		return ErrRoleNotFound
	} else if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}

	query := "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2"
	result, err := r.db.Exec(ctx, query, userID, roleID)
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
	query := `
		SELECT id, name, description, parent, created_at
		FROM roles
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
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

	return roles, nil
}

func (r *Repository) GetRoleHierarchy(ctx context.Context) (map[string][]string, error) {
	query := `
		SELECT parent_role, child_role
		FROM role_hierarchy
		ORDER BY parent_role, child_role
	`

	rows, err := r.db.Query(ctx, query)
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

	return hierarchy, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]UserRoles, error) {
	query := `
		SELECT
			ur.user_id,
			MAX(ur.assigned_at) as updated_at,
			ARRAY_AGG(r.name ORDER BY r.name) as roles
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		GROUP BY ur.user_id
		ORDER BY ur.user_id
	`

	rows, err := r.db.Query(ctx, query)
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

	return users, nil
}

func (r *Repository) CreateRole(ctx context.Context, name, description, parent string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO roles (id, name, description, parent)
		VALUES ($1, $2, $3, NULLIF($4, ''))
	`

	_, err = tx.Exec(ctx, query, uuid.New().String(), name, description, parent)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	if parent != "" {
		hierQuery := `
			INSERT INTO role_hierarchy (parent_role, child_role)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err = tx.Exec(ctx, hierQuery, parent, name)
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
	query := "DELETE FROM roles WHERE name = $1"
	result, err := r.db.Exec(ctx, query, name)
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
	query := `
		INSERT INTO rbac_audit_log (user_id, action, role, actor, success, error)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
	`

	_, err := r.db.Exec(ctx, query, userID, action, role, actor, success, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to log audit entry: %w", err)
	}

	return nil
}

func (r *Repository) GetAuditLog(ctx context.Context, limit int) ([]AuditEntry, error) {
	query := `
		SELECT id, timestamp, user_id, action, role, actor, success, COALESCE(error, '') as error
		FROM rbac_audit_log
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
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

	return entries, nil
}
