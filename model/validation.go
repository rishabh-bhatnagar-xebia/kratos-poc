package model

import (
	"errors"
	"fmt"
)

func (group *Group) Validate() error {
	switch *group {
	case VIEWER, ADMIN, MAKER, REVIEWER:
		return nil
	default:
		return fmt.Errorf("invalid group type: %s", *group)
	}
}

func (si *SessionInfo) OfGroup(target Group) bool {
	for _, group := range si.Identity.PublicMetadata.Groups {
		if group == target {
			return true
		}
	}
	return false
}

func (si *SessionInfo) Validate() error {
	return validateGroups(si.Identity.PublicMetadata.Groups)
}

func validateGroups(groups []Group) error {
	for _, group := range groups {
		err := group.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (user *User) Validate() error {
	if len(user.Groups) < 1 {
		return errors.New("there must be at least one group")
	}
	return validateGroups(user.Groups)
}
