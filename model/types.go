package model

import (
	"encoding/json"
)

type Group string

const (
	VIEWER   Group = "viewer"
	ADMIN          = "admin"
	MAKER          = "maker"
	REVIEWER       = "reviewer"
)

type User struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	IsActive bool    `json:"isActive"`
	Groups   []Group `json:"roles"`
}

type SessionInfo struct {
	Identity struct {
		ID             string `json:"id"`
		PublicMetadata struct {
			Groups []Group `json:"groups"`
		} `json:"metadata_public"`
	} `json:"identity"`
}

type UserSchemaState string

const (
	ACTIVE   UserSchemaState = "active"
	INACTIVE                 = "inactive"
)

type UserSchema struct {
	ID             string             `json:"id"`
	Traits         UserTrait          `json:"traits"`
	MetadataPubilc UserPublicMetadata `json:"metadata_public"`
	Creds          Credentials        `json:"credentials"`
	State          UserSchemaState    `json:"state"`
}

func (us *UserSchema) MarshalJSON() ([]byte, error) {
	// skips the hidden attributes like password before converting to json
	type userSchemaOut struct {
	}

	var isActive bool
	if us.State == ACTIVE {
		isActive = true
	} else {
		isActive = false
	}
	out := struct {
		Username string  `json:"username"`
		IsActive bool    `json:"isActive"`
		Groups   []Group `json:"roles"`
		ID       string  `json:"id"`
	}{
		ID:       us.ID,
		IsActive: isActive,
		Username: us.Traits.Username,
		Groups:   us.MetadataPubilc.Groups,
	}
	if len(out.Groups) == 0 {
		out.Groups = make([]Group, 0)
	}
	return json.Marshal(out)
}

type Credentials struct {
	Password CredentialPassword `json:"password"`
}

type CredentialPassword struct {
	Config UserPassword `json:"config"`
}

type UserPassword struct {
	Password       string `json:"password,omitempty"`
	HashedPassword string `json:"hashed_password,omitempty"`
}

type UserPublicMetadata struct {
	Groups []Group `json:"groups"`
}

type UserTrait struct {
	Username string `json:"username"`
}
