package rbac

import (
	rbacMigrations "github.com/nicolasbonnici/gorest-rbac/migrations"
	"github.com/nicolasbonnici/gorest/migrations"
)

func GetMigrations() migrations.MigrationSource {
	return rbacMigrations.GetMigrations()
}
