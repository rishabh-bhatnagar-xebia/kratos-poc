package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kratos-rbac/model"
	"net/http"
	"strings"
)

const KRATOS_ADMIN_URI string = "http://127.0.0.1:4434"

func setJson(w http.ResponseWriter) {
	// set content type to JSON
	w.Header().Set("Content-Type", "application/json")
}

func getID(r *http.Request) (id string, has bool) {
	urlParts := strings.Split(r.URL.String(), "/")
	if len(urlParts) < 2 {
		return "", false
	}
	return urlParts[len(urlParts)-1], true

	// use below if id is given in ?id= form
	//if !r.URL.Query().Has("id") {
	//	return "", false
	//}
	//return r.URL.Query().Get("id"), true
}

func validateID(id string) error {
	if strings.ContainsRune(id, '?') {
		return errors.New("id cannot have non-alphanumeric chars")
	}
	if len(id) == 0 {
		return errors.New("id cannot be empty")
	}
	return nil
}

func validateAndGetUser(w http.ResponseWriter, r *http.Request) (user *model.UserDetails, err error) {
	setJson(w)

	// decode the request body into a UserDetails struct
	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate the user details
	err = user.Validate()
	if err != nil {
		return user, err
	}

	return
}

// for all attributes
func validateAndGetUserAll(w http.ResponseWriter, r *http.Request) (user *model.UserDetails, err error) {
	user, err = validateAndGetUser(w, r)
	if err != nil {
		return
	}
	if len(user.Username) == 0 {
		// return the partially filled data in the user
		return user, errors.New("expected a valid username")
	}
	return user, nil
}

func validateAndGetID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, present := getID(r)
	if !present {
		http.Error(w, "missing param id", http.StatusBadRequest)
		return "", false
	}
	err := validateID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", false
	}
	return id, true
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := validateAndGetUserAll(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userJson, err := json.Marshal(model.UserSchema{
		Traits: model.UserTrait{
			Username: user.Username,
		},
		MetadataPubilc: model.UserPublicMetadata{
			Roles: user.Roles,
		},
		Creds: model.Credentials{Password: model.CredentialPassword{Config: model.UserPassword{Password: "password"}}},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	resp, err := requestJsonWithCookies(
		http.MethodPost,
		KRATOS_ADMIN_URI+"/admin/identities",
		bytes.NewReader(userJson),
		r.Cookies(),
	)
	b, err := io.ReadAll(resp.Body)

	var si model.SessionInfo
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&si)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(b)
}

func requestJsonWithCookies(method string, url string, body io.Reader, cookies []*http.Cookie) (*http.Response, error) {
	client := http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return client.Do(req)

}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("inside update user")
	user, err := validateAndGetUser(w, r)
	if err != nil {
		return
	}

	id, valid := validateAndGetID(w, r)
	if !valid {
		return
	}

	fmt.Printf("user found %+v", user)

	w.Write([]byte("will update person with id: " + id))

	var patch Patch
	if len(user.Username) != 0 {
		patch = append(patch, PatchRequest{
			Operation: REPLACE,
			Path:      "/username",
			Value:     user.Username,
		})
	}
	sb, err := io.ReadAll(r.Body)
	if strings.Contains(string(sb), "isActive") {
		patch = append(patch, PatchRequest{
			Operation: REPLACE,
			Path:      "/state",
			Value:     user.IsActive,
		})
	}
	if len(user.Roles) == 0 {
		patch = append(patch, PatchRequest{
			Operation: REPLACE,
			Path:      "/metadata_public",
			Value:     user.Roles,
		})
	}

	content, err := GetPatchBody(patch)
	fmt.Println("content", string(content), err)
	resp, err := requestJsonWithCookies(http.MethodPatch, KRATOS_ADMIN_URI+"/admin/identities/"+id, bytes.NewReader(content), r.Cookies())
	content, err = io.ReadAll(resp.Body)
	w.Write(content)
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := requestJsonWithCookies(http.MethodGet, KRATOS_ADMIN_URI+"/admin/identities", nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // 400 or 500?
	}
	content, _ := io.ReadAll(resp.Body)
	fmt.Println(string(content))
	var users []model.UserSchema
	err = json.Unmarshal(content, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("%+v\n", users)

	content, err = json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setJson(w)
	w.Write(content)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In delete user")
	id, valid := validateAndGetID(w, r)
	if !valid {
		return
	}

	fmt.Println("user with id=" + id + " will be deleted\n")
	// todo: delete user with id=$id
	fmt.Printf("deleting user with user id %s\n", id)
	resp, err := requestJsonWithCookies(http.MethodDelete, KRATOS_ADMIN_URI+"/admin/identities/"+id, strings.NewReader(""), r.Cookies())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500 or 400?
		return
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error occured while deleting")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode == http.StatusNoContent {
		w.Write([]byte("user deleted successfully!"))
	} else {
		http.Error(w, string(rb), http.StatusBadRequest)
	}
}

func endpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serving user request", r.Method)
	switch r.Method {
	case http.MethodPost:
		CreateUser(w, r)
	case http.MethodPut, http.MethodPatch:
		UpdateUser(w, r)
	case http.MethodGet:
		ListUsers(w, r)
	case http.MethodDelete:
		DeleteUser(w, r)
	default:
		w.Write([]byte("oh no, unhandled endpoint found: " + r.Method))
	}
	w.Write([]byte("\n"))
}

//	func AddRoutes(router *http.ServeMux, ms []model.Middleware, app model.App) {
//		commonMiddleware := ms[0]
//		for i := 1; i < len(ms); i++ {
//			commonMiddleware = func(handlerFunc http.HandlerFunc) http.HandlerFunc {
//				return commonMiddleware(ms[i](handlerFunc))
//			}
//		}
//		router.HandleFunc("/v1/user/", commonMiddleware(endpoint(app)))
//	}
func AddRoutes(router *http.ServeMux, ms Middleware) {
	router.HandleFunc("/v1/user/", RequestLogger(ms(endpoint)))
}
