package rbac

func GetMigrations() map[string]string {
	return map[string]string{
		"001_create_roles_table": `
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_parent FOREIGN KEY (parent) REFERENCES roles(name) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_parent ON roles(parent);
`,
		"002_create_user_roles_table": `
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(255),
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
`,
		"003_create_role_hierarchy_table": `
CREATE TABLE IF NOT EXISTS role_hierarchy (
    parent_role VARCHAR(100) NOT NULL,
    child_role VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (parent_role, child_role),
    CONSTRAINT fk_parent_role FOREIGN KEY (parent_role) REFERENCES roles(name) ON DELETE CASCADE,
    CONSTRAINT fk_child_role FOREIGN KEY (child_role) REFERENCES roles(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_role_hierarchy_parent ON role_hierarchy(parent_role);
CREATE INDEX IF NOT EXISTS idx_role_hierarchy_child ON role_hierarchy(child_role);
`,
		"004_create_audit_log_table": `
CREATE TABLE IF NOT EXISTS rbac_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    role VARCHAR(100) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON rbac_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON rbac_audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON rbac_audit_log(action);
`,
		"005_insert_default_roles": `
INSERT INTO roles (name, description) VALUES ('admin', 'System administrator with full access')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description) VALUES ('user', 'Basic authenticated user')
ON CONFLICT (name) DO NOTHING;
`,
	}
}
