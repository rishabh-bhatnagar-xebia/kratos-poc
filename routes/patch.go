package routes

import (
	"encoding/json"
)

type PatchOperation string

const (
	REPLACE PatchOperation = "replace"
	ADD                    = "add"
	REMOVE                 = "remove"
)

type PatchRequest struct {
	Operation PatchOperation `json:"op"`
	Path      string         `json:"path"`

	// Value will be empty for REPLACE operation
	Value any `json:"value,omitempty"`
}

type Patch []PatchRequest

func GetPatchBody(patch Patch) ([]byte, error) {
	content, err := json.Marshal(patch)
	return content, err
}
