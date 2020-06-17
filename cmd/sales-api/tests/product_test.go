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

// TestProducts runs a series of tests to exercise Product behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestProducts(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := ProductTests{
		app:       handlers.API("develop", shutdown, test.Log, test.DB, test.Authenticator),
		userToken: test.Token("admin@example.com", "gophers"),
	}

	// ADMIN role
	t.Run("postProduct400", tests.postProduct400)
	t.Run("postProduct401", tests.postProduct401)
	t.Run("getProduct404", tests.getProduct404)
	t.Run("getProduct400", tests.getProduct400)
	t.Run("deleteProductNotFound", tests.deleteProductNotFound)
	t.Run("putProduct404", tests.putProduct404)
	t.Run("crudProductAdmin", tests.crudProductAdmin)

	// USER role
	tests.userToken = test.Token("user@example.com", "gophers")
	t.Run("postProduct401", tests.postProduct401)
	t.Run("putProduct404", tests.putProduct404)
	t.Run("crudProductUser", tests.crudProductUser)
}

// ProductTests holds methods for each product subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type ProductTests struct {
	app       http.Handler
	userToken string
}

// postProduct400 validates a product can't be created with the endpoint
// unless a valid product document is submitted.
func (pt *ProductTests) postProduct400(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new product can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete product value.", testID)
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
					{Field: "cost", Error: "cost is a required field"},
					{Field: "quantity", Error: "quantity must be 1 or greater"},
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

// postProduct401 validates a product can't be created with the endpoint
// unless the user is authenticated
func (pt *ProductTests) postProduct401(t *testing.T) {
	np := data.NewProduct{
		Name:     "Comic Books",
		Cost:     25,
		Quantity: 60,
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting an authorization header

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new product can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete product value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// getProduct400 validates a product request for a malformed id.
func (pt *ProductTests) getProduct400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with a malformed id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
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

// getProduct404 validates a product request for a product that does not exist with the endpoint.
func (pt *ProductTests) getProduct404(t *testing.T) {
	id := "a224a8d6-3f9e-4b11-9900-e81a25d80702"

	r := httptest.NewRequest(http.MethodGet, "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := "Product not found"
			if !strings.Contains(recv, resp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// deleteProductNotFound validates deleting a product that does not exist is not a failure.
func (pt *ProductTests) deleteProductNotFound(t *testing.T) {
	id := "112262f1-1a77-4374-9f22-39e575aa6348"

	r := httptest.NewRequest(http.MethodDelete, "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a product that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// putProduct404 validates updating a product that does not exist.
func (pt *ProductTests) putProduct404(t *testing.T) {
	up := data.UpdateProduct{
		Name: tests.StringPointer("Nonexistent"),
	}

	id := "9b468f90-1cf1-4377-b3fa-68b450d632a0"

	body, err := json.Marshal(&up)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/products/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a product that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			recv := w.Body.String()
			resp := "Product not found"
			if !strings.Contains(recv, resp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// crudProductAdmin performs a complete test of CRUD against the api.
func (pt *ProductTests) crudProductAdmin(t *testing.T) {
	p := pt.postProduct201(t)
	defer pt.deleteProduct204(t, p.ID)

	pt.getProduct200(t, p.ID)
	pt.putProduct204(t, p.ID)

	pt.postProductSale201(t, p.ID)
	pt.getProductSales200(t, p.ID)
}

// crudProductUser performs tests for produtct creation and update
// as well as creation of a sale.
func (pt *ProductTests) crudProductUser(t *testing.T) {
	p := pt.postProduct201(t)
	pt.getProduct200(t, p.ID)
	pt.putProduct204(t, p.ID)
	pt.postProductSale403(t, p.ID)
}

// postProduct201 validates a product can be created with the endpoint.
func (pt *ProductTests) postProduct201(t *testing.T) data.Product {
	np := data.NewProduct{
		Name:     "Comic Books",
		Cost:     25,
		Quantity: 60,
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	// p is the value we will return.
	var p data.Product

	t.Log("Given the need to create a new product with the products endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared product value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy p.
			want := p
			want.Name = "Comic Books"
			want.Cost = 25
			want.Quantity = 60
			want.Revenue = 0
			want.Sold = 0

			if diff := cmp.Diff(want, p); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}

	return p
}

// deleteProduct200 validates deleting a product that does exist.
func (pt *ProductTests) deleteProduct204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a product that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// getProduct200 validates a product request for an existing id.
func (pt *ProductTests) getProduct200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/products/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a product that exists.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new product %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var p data.Product
			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			want := p
			want.ID = id
			want.Name = "Comic Books"
			want.Cost = 25
			want.Quantity = 60
			want.Revenue = 0
			want.Sold = 0

			if diff := cmp.Diff(want, p); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putProduct204 validates updating a product that does exist.
func (pt *ProductTests) putProduct204(t *testing.T, id string) {
	body := `{"name": "Graphic Novels", "cost": 100}`
	r := httptest.NewRequest(http.MethodPut, "/v1/products/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to update a product with the products endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified product value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/products/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)

			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ru data.Product
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ru.Name != "Graphic Novels" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", tests.Failed, testID, ru.Name, "Graphic Novels")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)
		}
	}
}

// postProductSale403 validates that creating sales
// with role=USER is forbidden.
func (pt *ProductTests) postProductSale403(t *testing.T, id string) {

	ns := data.NewSale{
		Quantity: 3,
		Paid:     70,
	}

	body, err := json.Marshal(&ns)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/products/"+id+"/sales", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to create product sales.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared sales value.", testID)
		{
			if w.Code != http.StatusForbidden {
				t.Logf("\t%s\tTest %d:\tShould NOT be able to add a new sale with role=%s. : %v", tests.Failed, testID, auth.RoleUser, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to add a new sale with role=%s.", tests.Success, testID, auth.RoleUser)

			recv := w.Body.String()
			resp := `{"error":"you are not authorized for that action"}`
			if resp != recv {
				t.Logf("\t\tTest %d:\tGot : %v", testID, recv)
				t.Logf("\t\tTest %d:\tWant: %v", testID, resp)
				t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// postProductSale201 validates sales can be created with the endpoint.
func (pt *ProductTests) postProductSale201(t *testing.T, id string) {

	ns := data.NewSale{
		Quantity: 3,
		Paid:     70,
	}

	body, err := json.Marshal(&ns)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/products/"+id+"/sales", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to create product sales.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared sales value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Logf("\t%s\tTest %d:\tShould be able to add a new sale. : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to add a new sale.", tests.Success, testID)

			var s data.Sale
			if err := json.NewDecoder(w.Body).Decode(&s); err != nil {
				t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy s.
			want := s
			want.ProductID = id
			want.Quantity = ns.Quantity
			want.Paid = ns.Paid

			if diff := cmp.Diff(want, s); diff != "" {
				t.Logf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// getProductSales200 validates sales request for a product.
func (pt *ProductTests) getProductSales200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/products/"+id+"/sales", nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)

	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate sales.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the product %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var sales []data.Sale
			if err := json.NewDecoder(w.Body).Decode(&sales); err != nil {
				t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if exp, got := 1, len(sales); exp != got {
				t.Logf("\t%s\tTest %d:\tShould get back [%v] sales for the product, got %v", tests.Failed, testID, exp, got)
			}
			t.Logf("\t%s\tTest %d:\tShould get back ONE sale for the product %v", tests.Success, testID, sales[0])

			// get the product and compare sales values
			r := httptest.NewRequest(http.MethodGet, "/v1/products/"+id, nil)
			w := httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)

			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var p data.Product
			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy sales[0].
			want := sales[0]
			want.ProductID = p.ID
			want.Quantity = p.Sold
			want.Paid = p.Revenue

			if diff := cmp.Diff(want, sales[0]); diff != "" {
				t.Logf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}
