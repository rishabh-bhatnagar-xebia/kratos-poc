package model

import (
	"errors"
	"fmt"
	"strings"
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
	if len(user.Username) == 0 {
		return errors.New("expected a valid username")
	}
	return validateRoles(user.Roles)
}

func ValidateID(id string) error {
	if strings.ContainsRune(id, '?') {
		return errors.New("id cannot have non-alphanumeric chars")
	}
	if len(id) == 0 {
		return errors.New("id cannot be empty")
	}
	return nil
}
