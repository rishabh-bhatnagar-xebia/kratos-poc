package routes

import (
	"encoding/json"
	"errors"
	"io"
	"kratos-rbac/model"
	"kratos-rbac/utils"
	"net/http"
	"strings"
)

func readID(r *http.Request) (id string, err error) {
	urlParts := strings.Split(r.URL.String(), "/")
	if len(urlParts) < 2 {
		return "", errors.New("missing param id in the request url")
	}
	return urlParts[len(urlParts)-1], nil
}

func readUser(r *http.Request) (user *model.UserDetails, err error) {
	// decode the request body into a UserDetails struct
	err = json.NewDecoder(r.Body).Decode(&user)
	return user, err
}

func readAndValidateId(r *http.Request) (string, error) {
	id, err := readID(r)
	if err != nil {
		return "", err
	}
	err = model.ValidateID(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := readUser(r)
	if err == nil {
		err = user.Validate()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = model.CreateUser(user, r.Cookies())

	if err == nil {
		// user created successfully
		w.WriteHeader(201)
		return
	}

	// echo the json error message from kratos verbatim
	utils.SetJson(w)
	w.Write([]byte(err.Error()))
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	userDetails, err := model.FetchUserDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // 400 or 500?
	}

	utils.SetJson(w)
	content, err := json.Marshal(userDetails)
	w.Write(content)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	user, err := readUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := readAndValidateId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rb, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = model.UpdateUser(rb, user, id, r.Cookies())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := readAndValidateId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err, statusCode := model.DeleteUser(id, r.Cookies())
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.WriteHeader(statusCode)
}
