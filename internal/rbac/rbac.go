package rbac

import (
	_ "embed"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

//go:embed model.conf
var modelConf string

// NewEnforcer builds a Casbin enforcer backed by the app database.
func NewEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "casbin_rule", "")
	if err != nil {
		return nil, err
	}
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}
	return casbin.NewEnforcer(m, adapter)
}

// SeedDefaultPolicies ensures base role permissions exist (idempotent).
func SeedDefaultPolicies(e *casbin.Enforcer) error {
	_, err := e.AddPolicy("admin", "/kong-admin/*", "*")
	if err != nil {
		return err
	}
	_, _ = e.AddPolicy("viewer", "/kong-admin/*", "GET")
	return nil
}

// SyncUserGroups updates Casbin grouping from the relational model.
func SyncUserGroups(e *casbin.Enforcer, username string, groupNames []string) error {
	_, err := e.RemoveFilteredGroupingPolicy(0, username)
	if err != nil {
		return err
	}
	for _, g := range groupNames {
		if _, err := e.AddGroupingPolicy(username, g); err != nil {
			return err
		}
	}
	return nil
}
