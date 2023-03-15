package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"kratos-rbac/constants"
	"kratos-rbac/routes"
	"kratos-rbac/utils"
	"net/http"
	"strings"
)

func FetchUserDetails() ([]FilteredUserSchema, error) {
	resp, err := utils.RequestJsonWithCookies(http.MethodGet, constants.URL_KRATOS_IDENTITIES, nil, nil)
	if err != nil {
		return nil, err
	}

	content, _ := io.ReadAll(resp.Body)
	var users []KratosUserSchema
	err = json.Unmarshal(content, &users)
	if err != nil {
		return nil, err
	}

	var filteredUserDetails []FilteredUserSchema
	for _, user := range users {
		filteredUserDetails = append(filteredUserDetails, user.Filter())
	}
	return filteredUserDetails, nil
}

func CreateUser(user *UserDetails, cookies []*http.Cookie) error {
	userJson, err := json.Marshal(KratosUserSchema{
		Traits: UserTrait{
			Username: user.Username,
		},
		MetadataPubilc: UserPublicMetadata{
			Roles: user.Roles,
		},
		Creds: Credentials{Password: CredentialPassword{Config: UserPassword{Password: "password"}}},
	})
	if err != nil {
		return err
	}

	resp, err := utils.RequestJsonWithCookies(
		http.MethodPost,
		constants.URL_KRATOS_IDENTITIES,
		bytes.NewReader(userJson),
		cookies,
	)
	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	// error has occurred creating a user
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return errors.New(string(b))
}

func UpdateUser(requestContent []byte, user *UserDetails, id string, cookies []*http.Cookie) error {
	var patch routes.Patch
	if len(user.Username) != 0 {
		patch = append(patch, routes.PatchRequest{
			Operation: routes.REPLACE,
			Path:      "/username",
			Value:     user.Username,
		})
	}
	if strings.Contains(string(requestContent), "isActive") {
		patch = append(patch, routes.PatchRequest{
			Operation: routes.REPLACE,
			Path:      "/state",
			Value:     user.IsActive,
		})
	}
	if len(user.Roles) == 0 {
		patch = append(patch, routes.PatchRequest{
			Operation: routes.REPLACE,
			Path:      "/metadata_public",
			Value:     user.Roles,
		})
	}

	content, err := routes.GetPatchBody(patch)
	if err != nil {
		return err
	}
	resp, err := utils.RequestJsonWithCookies(http.MethodPatch, constants.URL_KRATOS_IDENTITIES+id, bytes.NewReader(content), cookies)
	content, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(string(content))
}

func DeleteUser(id string, cookies []*http.Cookie) (error, int) {
	resp, err := utils.RequestJsonWithCookies(http.MethodDelete, constants.URL_KRATOS_IDENTITIES+id, strings.NewReader(""), cookies)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil, http.StatusNoContent
	}
	return errors.New(string(rb)), http.StatusBadRequest
}
