package rbac

import "github.com/casbin/casbin/v2"

// SetRolePolicies replaces all p-policies for a role (first column = role name).
func SetRolePolicies(e *casbin.Enforcer, role string, objectAction [][2]string) error {
	if _, err := e.RemoveFilteredPolicy(0, role); err != nil {
		return err
	}
	for _, row := range objectAction {
		if len(row) != 2 {
			continue
		}
		obj, act := row[0], row[1]
		if obj == "" || act == "" {
			continue
		}
		if _, err := e.AddPolicy(role, obj, act); err != nil {
			return err
		}
	}
	return nil
}

// RenameRole rewrites Casbin p and g rows from oldName to newName.
func RenameRole(e *casbin.Enforcer, oldName, newName string) error {
	if oldName == newName {
		return nil
	}
	pols, err := e.GetFilteredPolicy(0, oldName)
	if err != nil {
		return err
	}
	if _, err := e.RemoveFilteredPolicy(0, oldName); err != nil {
		return err
	}
	for _, row := range pols {
		if len(row) < 3 {
			continue
		}
		if _, err := e.AddPolicy(newName, row[1], row[2]); err != nil {
			return err
		}
	}
	groups, err := e.GetFilteredGroupingPolicy(1, oldName)
	if err != nil {
		return err
	}
	if _, err := e.RemoveFilteredGroupingPolicy(1, oldName); err != nil {
		return err
	}
	for _, row := range groups {
		if len(row) < 2 {
			continue
		}
		if _, err := e.AddGroupingPolicy(row[0], newName); err != nil {
			return err
		}
	}
	return nil
}

// RemoveRoleFromCasbin drops all p and g rows for this role name.
func RemoveRoleFromCasbin(e *casbin.Enforcer, role string) error {
	if _, err := e.RemoveFilteredPolicy(0, role); err != nil {
		return err
	}
	if _, err := e.RemoveFilteredGroupingPolicy(1, role); err != nil {
		return err
	}
	return nil
}

func IsSystemRole(name string) bool {
	return name == "admin" || name == "viewer"
}
