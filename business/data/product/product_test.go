package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/product"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/business/tests"
)

func TestProduct(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	p := product.New(log, db)

	t.Log("Given the need to work with Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
		{
			ctx := context.Background()
			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e",
					Audience:  jwt.ClaimStrings{"students"},
					ExpiresAt: jwt.At(now.Add(time.Hour)),
					IssuedAt:  jwt.At(now),
				},
				Roles: []string{auth.RoleAdmin, auth.RoleUser},
			}

			np := product.NewProduct{
				Name:     "Comic Books",
				Cost:     10,
				Quantity: 55,
			}

			prd, err := p.Create(ctx, traceID, claims, np, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a product.", tests.Success, testID)

			saved, err := p.QueryByID(ctx, traceID, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", tests.Success, testID)

			if diff := cmp.Diff(prd, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd := product.UpdateProduct{
				Name:     tests.StringPointer("Comics"),
				Cost:     tests.IntPointer(50),
				Quantity: tests.IntPointer(40),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := p.Update(ctx, traceID, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update product.", tests.Success, testID)

			saved, err = p.QueryByID(ctx, traceID, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			// Check specified fields were updated. Make a copy of the original product
			// and change just the fields we expect then diff it with what was saved.
			want := prd
			want.Name = *upd.Name
			want.Cost = *upd.Cost
			want.Quantity = *upd.Quantity
			want.DateUpdated = updatedTime

			if diff := cmp.Diff(want, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

			upd = product.UpdateProduct{
				Name: tests.StringPointer("Graphic Novels"),
			}

			if err := p.Update(ctx, traceID, claims, prd.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", tests.Success, testID)

			saved, err = p.QueryByID(ctx, traceID, prd.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", tests.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", tests.Success, testID)
			}

			if err := p.Delete(ctx, traceID, claims, prd.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete product.", tests.Success, testID)

			_, err = p.QueryByID(ctx, traceID, prd.ID)
			if errors.Cause(err) != product.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", tests.Success, testID)
		}
	}
}

func TestProductPaging(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	schema.Seed(db)

	p := product.New(log, db)

	t.Log("Given the need to page through Product records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 products.", testID)
		{
			ctx := context.Background()
			traceID := "00000000-0000-0000-0000-000000000000"

			pageNumber := 1
			rowsPerPage := 1
			products1, err := p.Query(ctx, traceID, pageNumber, rowsPerPage)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve products for page 1 : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve products for page 1.", tests.Success, testID)

			if len(products1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single product.", tests.Success, testID)

			pageNumber = 2
			products2, err := p.Query(ctx, traceID, pageNumber, rowsPerPage)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve products for page 2 : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve products for page 2.", tests.Success, testID)

			if len(products2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single product : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single product.", tests.Success, testID)

			if products1[0].ID == products2[0].ID {
				t.Logf("\t\tTest %d:\tProduct1: %v", testID, products1[0].ID)
				t.Logf("\t\tTest %d:\tProduct2: %v", testID, products2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different products : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different products.", tests.Success, testID)
		}
	}
}
