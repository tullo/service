package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tullo/service/cmd/sales-api/internal/handlers"
	"github.com/tullo/service/internal/data"
	"github.com/tullo/service/internal/platform/auth"
	"github.com/tullo/service/internal/platform/tests"
	"github.com/tullo/service/internal/platform/web"
)

// TestUsers is the entry point for testing user management functions.
func TestUsers(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := UserTests{
		app:        handlers.API("develop", shutdown, test.Log, test.DB, test.Authenticator),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("getToken401", tests.getToken401)
	t.Run("getToken200", tests.getToken200)
	t.Run("postUser400", tests.postUser400)
	t.Run("postUser401", tests.postUser401)
	t.Run("postUser403", tests.postUser403)
	t.Run("getUser400", tests.getUser400)
	t.Run("getUser403", tests.getUser403)
	t.Run("getUser404", tests.getUser404)
	t.Run("deleteUserNotFound", tests.deleteUserNotFound)
	t.Run("putUser404", tests.putUser404)
	t.Run("crudUsers", tests.crudUser)
}

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// getToken401 ensures an unknown user can't generate a token.
func (ut *UserTests) getToken401(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("unknown@example.com", "some-password")

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to deny tokens to unknown users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with an unrecognized email.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// getToken200
func (ut *UserTests) getToken200(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@example.com", "gophers")

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to issues tokens to known users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with valid credentials.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response.", tests.Success, testID)

			// TODO(jlw) Should we ensure the token is valid?
		}
	}
}

// postUser400 validates a user can't be created with the endpoint
// unless a valid user document is submitted.
func (ut *UserTests) postUser400(t *testing.T) {
	body, err := json.Marshal(&data.NewUser{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", tests.Success, testID)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type.", tests.Success, testID)

			// Define what we want to see.
			want := web.ErrorResponse{
				Error: "field validation error",
				Fields: []web.FieldError{
					{Field: "name", Error: "name is a required field"},
					{Field: "email", Error: "email is a required field"},
					{Field: "roles", Error: "roles is a required field"},
					{Field: "password", Error: "password is a required field"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(want, got, sorter); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// postUser401 validates a user can't be created unless the calling user is
// authenticated.
func (ut *UserTests) postUser401(t *testing.T) {
	body, err := json.Marshal(&data.User{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.userToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", tests.Success, testID)
		}
	}
}

// postUser403 validates a user can't be created unless the calling user is
// an admin user. Regular users can't do this.
func (ut *UserTests) postUser403(t *testing.T) {
	body, err := json.Marshal(&data.User{})
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting the Authorization header

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete user value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// getUser400 validates a user request for a malformed userid.
func (ut *UserTests) getUser400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with a malformed userid.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := `{"error":"ID is not in its proper form"}`
			if resp != recv {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// getUser403 validates a regular user can't fetch anyone but themselves
func (ut *UserTests) getUser403(t *testing.T) {
	t.Log("Given the need to validate regular users can't fetch other users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching the admin user as a regular user.", testID)
		{
			r := httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.AdminID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.userToken)

			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := `{"error":"attempted action is not allowed"}`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}

		testID = 1
		t.Logf("\tTest %d:\tWhen fetching the user as themselves.", testID)
		{

			r := httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.UserID, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.userToken)

			ut.app.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)
		}
	}
}

// getUser404 validates a user request for a user that does not exist with the endpoint.
func (ut *UserTests) getUser404(t *testing.T) {
	id := "c50a5d66-3c4d-453f-af3f-bc960ed1a503"

	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := "not found"
			if !strings.Contains(recv, resp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// deleteUserNotFound validates deleting a user that does not exist is not a failure.
func (ut *UserTests) deleteUserNotFound(t *testing.T) {
	id := "a71f77b2-b1ae-4964-a847-f9eecba09d74"

	r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// putUser404 validates updating a user that does not exist.
func (ut *UserTests) putUser404(t *testing.T) {
	u := data.UpdateUser{
		Name: tests.StringPointer("Doesn't Exist"),
	}

	id := "3097c45e-780a-421b-9eae-43c2fda2bf14"

	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a user that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := "not found"
			if !strings.Contains(recv, resp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// crudUser performs a complete test of CRUD against the api.
func (ut *UserTests) crudUser(t *testing.T) {
	nu := ut.postUser201(t)
	defer ut.deleteUser204(t, nu.ID)

	ut.getUser200(t, nu.ID)
	ut.putUser204(t, nu.ID)
	ut.putUser403(t, nu.ID)
}

// postUser201 validates a user can be created with the endpoint.
func (ut *UserTests) postUser201(t *testing.T) data.User {
	nu := data.NewUser{
		Name:            "Bill Kennedy",
		Email:           "bill@ardanlabs.com",
		Roles:           []string{auth.RoleAdmin},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	body, err := json.Marshal(&nu)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	// u is the value we will return.
	var u data.User

	t.Log("Given the need to create a new user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared user value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy u.
			want := u
			want.Name = "Bill Kennedy"
			want.Email = "bill@ardanlabs.com"
			want.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(want, u); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}

	return u
}

// deleteUser200 validates deleting a user that does exist.
func (ut *UserTests) deleteUser204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a user that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// getUser200 validates a user request for an existing userid.
func (ut *UserTests) getUser200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a user that exsits.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new user %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var u data.User
			if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			want := u
			want.ID = id
			want.Name = "Bill Kennedy"
			want.Email = "bill@ardanlabs.com"
			want.Roles = []string{auth.RoleAdmin}

			if diff := cmp.Diff(want, u); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putUser204 validates updating a user that does exist.
func (ut *UserTests) putUser204(t *testing.T, id string) {
	body := `{"name": "Jacob Walker"}`

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.adminToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified user value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ut.adminToken)

			ut.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ru data.User
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ru.Name != "Jacob Walker" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", tests.Failed, testID, ru.Name, "Jacob Walker")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)

			if ru.Email != "bill@ardanlabs.com" {
				t.Fatalf("\t%s\tTest %d:\tShould not affect other fields like Email : got %q want %q", tests.Failed, testID, ru.Email, "bill@ardanlabs.com")
			}
			t.Logf("\t%s\tTest %d:\tShould not affect other fields like Email.", tests.Success, testID)
		}
	}
}

// putUser403 validates that a user can't modify users unless they are an admin.
func (ut *UserTests) putUser403(t *testing.T, id string) {
	body := `{"name": "Anna Walker"}`

	r := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ut.userToken)

	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to update a user with the users endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen a non-admin user makes a request", testID)
		{
			if w.Code != http.StatusForbidden {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 403 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 403 for the response.", tests.Success, testID)
		}
	}
}
