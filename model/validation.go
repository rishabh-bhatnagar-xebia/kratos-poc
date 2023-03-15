package model

import (
	"errors"
	"fmt"
)

func (group *Role) Validate() error {
	switch *group {
	case VIEWER, ADMIN, MAKER, REVIEWER:
		return nil
	default:
		return fmt.Errorf("invalid group type: %s", *group)
	}
}

func (si *SessionInfo) OfRole(target Role) bool {
	for _, group := range si.Identity.PublicMetadata.Roles {
		if group == target {
			return true
		}
	}
	return false
}

func (si *SessionInfo) Validate() error {
	return validateRoles(si.Identity.PublicMetadata.Roles)
}

func validateRoles(groups []Role) error {
	for _, group := range groups {
		err := group.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (user *UserDetails) Validate() error {
	if len(user.Roles) < 1 {
		return errors.New("there must be at least one group")
	}
	return validateRoles(user.Roles)
}
