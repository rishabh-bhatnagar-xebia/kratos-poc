package model

import "encoding/json"

type Role string

const (
	VIEWER   Role = "viewer"
	ADMIN         = "admin"
	MAKER         = "maker"
	REVIEWER      = "reviewer"
)

// UserDetails wraps information that user provides as an input
type UserDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsActive bool   `json:"isActive"`
	Roles    []Role `json:"roles"`
}

type SessionInfo struct {
	Identity struct {
		ID             string `json:"id"`
		PublicMetadata struct {
			Roles []Role `json:"groups"`
		} `json:"metadata_public"`
	} `json:"identity"`
}

type UserSchemaState string

const (
	ACTIVE   UserSchemaState = "active"
	INACTIVE                 = "inactive"
)

// KratosUserSchema represents User Schema struct of the identity as
// returned by kratos
type KratosUserSchema struct {
	ID             string             `json:"id,omitempty"`
	Traits         UserTrait          `json:"traits"`
	MetadataPubilc UserPublicMetadata `json:"metadata_public"`
	Creds          Credentials        `json:"credentials"`
	State          UserSchemaState    `json:"state,omitempty"`
}

// FilteredUserSchema is KratosUserSchema without sensitive values
type FilteredUserSchema struct {
	ID             string             `json:"id"`
	Traits         UserTrait          `json:"traits"`
	MetadataPubilc UserPublicMetadata `json:"metadata_public"`
	State          UserSchemaState    `json:"state,omitempty"`
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
	Roles []Role `json:"groups"`
}

type UserTrait struct {
	Username string `json:"username"`
}

// Filter excludes all the sensitive information
func (us *KratosUserSchema) Filter() FilteredUserSchema {
	return FilteredUserSchema{
		ID:             us.ID,
		Traits:         us.Traits,
		MetadataPubilc: us.MetadataPubilc,
		State:          us.State,
	}
}

func (fus *FilteredUserSchema) MarshalJSON() ([]byte, error) {
	var isActive bool
	if fus.State == ACTIVE {
		isActive = true
	} else {
		isActive = false
	}
	out := struct {
		Username string `json:"username"`
		IsActive bool   `json:"isActive"`
		Roles    []Role `json:"roles"`
		ID       string `json:"id"`
	}{
		ID:       fus.ID,
		IsActive: isActive,
		Username: fus.Traits.Username,
		Roles:    fus.MetadataPubilc.Roles,
	}
	if len(out.Roles) == 0 {
		out.Roles = make([]Role, 0)
	}
	return json.Marshal(out)
}
