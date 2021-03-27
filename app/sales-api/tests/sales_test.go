package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data/product"
	"github.com/tullo/service/business/data/sale"
	"github.com/tullo/service/business/tests"
)

func Test_Sales(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	defer teardown()

	p := product.New(log, db)
	s := sale.New(log, db)

	t.Log("Given the need to work with product Sales records.")

	now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)

	ctx := context.Background()

	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
			Audience:  jwt.ClaimStrings{"students"},
			ExpiresAt: jwt.At(now.Add(time.Hour)),
			IssuedAt:  jwt.At(now),
		},
		Roles: []string{auth.RoleAdmin, auth.RoleUser},
	}

	// Create two products to work with.
	newPuzzles := product.NewProduct{
		Name:     "Puzzles",
		Cost:     25,
		Quantity: 6,
	}

	traceID := "00000000-0000-0000-0000-000000000000"

	puzzles, err := p.Create(ctx, traceID, claims, newPuzzles, now)
	if err != nil {
		t.Fatalf("creating product: %s", err)
	}

	newToys := product.NewProduct{
		Name:     "Toys",
		Cost:     40,
		Quantity: 3,
	}
	toys, err := p.Create(ctx, traceID, claims, newToys, now)
	if err != nil {
		t.Fatalf("creating product: %s", err)
	}

	{ // Add and list

		testID := 0
		t.Logf("\tTest %d:\tWhen handling product Sales.", testID)

		ns := sale.NewSale{
			Quantity: 3,
			Paid:     70,
		}

		sale, err := s.AddSale(ctx, traceID, ns, puzzles.ID, now)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to add a new sale: %s", tests.Failed, testID, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to add a new sale.", tests.Success, testID)

		// Puzzles should show the 1 sale.
		sales, err := s.List(ctx, traceID, puzzles.ID)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tShould be able to list sales for a product: %s.", tests.Failed, testID, err)
		}
		t.Logf("\t%s\tTest %d:\tShould be able to list sales for a product.", tests.Success, testID)

		if exp, got := 1, len(sales); exp != got {
			t.Fatalf("\t%s\tTest %d:\tExpected sale list size %v, got %v", tests.Failed, testID, exp, got)
		}
		t.Logf("\t%s\tTest %d:\tShould get back ONE sale for the product", tests.Success, testID)

		if exp, got := sale.ID, sales[0].ID; exp != got {
			t.Fatalf("\t%s\tTest %d:\tExpected first sale ID %v, got %v", tests.Failed, testID, exp, got)
		}

		// Toys should have 0 sales.
		sales, err = s.List(ctx, traceID, toys.ID)
		if err != nil {
			t.Fatalf("\t%s\tTest %d:\tListing sales: %s", tests.Failed, testID, err)
		}
		if exp, got := 0, len(sales); exp != got {
			t.Fatalf("\t%s\tTest %d:\tExpected sale list size %v, got %v", tests.Failed, testID, exp, got)
		}
		t.Logf("\t%s\tTest %d:\tShould get back NO sales.", tests.Success, testID)
	}
}
