package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"kratos-rbac/constants"
	"kratos-rbac/model"
	"net/http"
	"strings"
)

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
}

func validateAndReadUser(r *http.Request) (user *model.UserDetails, err error) {
	// decode the request body into a UserDetails struct
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	// validate the user details
	err = user.Validate()
	if err != nil {
		return user, err
	}
	return
}

func validateAndReadID(r *http.Request) (string, error) {
	id, present := getID(r)
	if !present {
		return "", errors.New("missing param id")
	}
	err := model.ValidateID(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := validateAndReadUser(r)
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
		constants.URL_KRATOS_IDENTITIES,
		bytes.NewReader(userJson),
		r.Cookies(),
	)

	if resp.StatusCode == http.StatusCreated {
		w.WriteHeader(http.StatusCreated)
		return
	}

	b, err := io.ReadAll(resp.Body)
	// echo the json error message from kratos verbatim
	setJson(w)
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
	user, err := validateAndReadUser(r)
	if err != nil {
		return
	}

	id, err := validateAndReadID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	resp, err := requestJsonWithCookies(http.MethodPatch, constants.URL_KRATOS_IDENTITIES+id, bytes.NewReader(content), r.Cookies())
	content, err = io.ReadAll(resp.Body)
	w.Write(content)
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := requestJsonWithCookies(http.MethodGet, constants.URL_KRATOS_IDENTITIES, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // 400 or 500?
	}
	content, _ := io.ReadAll(resp.Body)
	var users []model.UserSchema
	err = json.Unmarshal(content, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content, err = json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setJson(w)
	w.Write(content)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := validateAndReadID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := requestJsonWithCookies(http.MethodDelete, constants.URL_KRATOS_IDENTITIES+id, strings.NewReader(""), r.Cookies())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500 or 400?
		return
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
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
