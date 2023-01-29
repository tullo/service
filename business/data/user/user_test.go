package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	// _ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tullo/service/business/auth"
	"github.com/tullo/service/business/data"
	"github.com/tullo/service/business/data/schema"
	"github.com/tullo/service/business/data/tests"
	"github.com/tullo/service/business/data/user"
)

func TestUser(t *testing.T) {
	//log, db, teardown := tests.NewUnit(t, tests.NewRoachDBSpec())
	//t.Cleanup(teardown)
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	log, db, teardown := tests.NewUnit(t, ctx)
	t.Cleanup(teardown)

	u := user.NewStore(log, db)

	t.Log("Given the need to work with User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "tullo@users.noreply.github.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := u.Create(ctx, traceID, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create user.", tests.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   usr.ID,
					Audience:  jwt.ClaimStrings{"students"},
					ExpiresAt: jwt.At(now.Add(time.Hour)),
					IssuedAt:  jwt.At(now),
				},
				Roles: []string{auth.RoleUser},
			}

			// Query own user while having USER authz.
			saved, err := u.QueryByID(ctx, traceID, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by ID.", tests.Success, testID)

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)

			upd := user.UpdateUser{
				Name:     tests.StringPointer("Andreas Amstutz"),
				Email:    tests.StringPointer("tullo@users.noreply.github.com"),
				Password: tests.StringPointer("gophercon-2021"),
				Roles:    []string{auth.RoleUser},
			}

			claims = auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Audience:  jwt.ClaimStrings{"students"},
					ExpiresAt: jwt.At(now.Add(time.Hour)),
					IssuedAt:  jwt.At(now),
				},
				Roles: []string{auth.RoleAdmin},
			}

			if err := u.Update(ctx, traceID, claims, usr.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update user.", tests.Success, testID)

			// Query updated user while having ADMIN authz.
			saved, err = u.QueryByEmail(ctx, traceID, claims, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by Email : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by Email.", tests.Success, testID)

			if saved.Name != *upd.Name {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Name.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Name)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", tests.Success, testID)
			}

			if saved.Email != *upd.Email {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Email.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Email)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email.", tests.Success, testID)
			}

			if saved.PasswordHash == usr.PasswordHash {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to PasswordHash.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.PasswordHash)
				t.Logf("\t\tTest %d:\tExp: %v", testID, usr.PasswordHash)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to PasswordHash.", tests.Success, testID)
			}

			if !(len(saved.Roles) == len(upd.Roles)) && saved.Roles[0] == upd.Roles[0] {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Roles.", tests.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Roles)
				t.Logf("\t\tTest %d:\tExp: %v", testID, upd.Roles)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Roles.", tests.Success, testID)
			}

			_, err = u.Create(ctx, traceID, nu, now)
			if errors.Cause(err) != data.ErrDuplicateEmail {
				t.Fatalf("\t%s\tShould not be able create user: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould not be able to create user.", tests.Success)

			if err := u.Delete(ctx, traceID, claims, "00000000-0000"); err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to delete user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to delete user.", tests.Success, testID)

			old := claims.Roles
			claims.Roles = []string{auth.RoleUser}
			if err := u.Delete(ctx, traceID, claims, usr.ID); err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to delete user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to delete user.", tests.Success, testID)

			claims.Roles = old
			if err := u.Delete(ctx, traceID, claims, usr.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete user.", tests.Success, testID)

			// Query deleted user while having ADMIN authz.
			_, err = u.QueryByID(ctx, traceID, claims, usr.ID)
			if errors.Cause(err) != data.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user.", tests.Success, testID)
		}
	}
}

func TestUserPaging(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	log, db, teardown := tests.NewUnit(t, ctx)
	t.Cleanup(teardown)

	schema.Seed(ctx, db)

	u := user.NewStore(log, db)

	t.Log("Given the need to page through User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 users.", testID)
		{
			ctx := context.Background()
			traceID := "00000000-0000-0000-0000-000000000000"

			users1, err := u.Query(ctx, traceID, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve users for page 1 : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve users for page 1.", tests.Success, testID)

			if len(users1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", tests.Success, testID)

			users2, err := u.Query(ctx, traceID, 2, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve users for page 2 : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve users for page 2.", tests.Success, testID)

			if len(users2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", tests.Success, testID)

			if users1[0].ID == users2[0].ID {
				t.Logf("\t\tTest %d:\tUser1: %v", testID, users1[0].ID)
				t.Logf("\t\tTest %d:\tUser2: %v", testID, users2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different users : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different users.", tests.Success, testID)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	log, db, teardown := tests.NewUnit(t, ctx)
	t.Cleanup(teardown)

	u := user.NewStore(log, db)

	t.Log("Given the need to authenticate users")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "tullo@users.noreply.github.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}

			usr, err := u.Create(ctx, traceID, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create user.", tests.Success, testID)

			claims, err := u.Authenticate(ctx, traceID, now, "tullo@users.noreply.github.com", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate claims : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate claims.", tests.Success, testID)

			want := auth.Claims{
				Roles: usr.Roles,
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   usr.ID,
					Audience:  jwt.ClaimStrings{"students"},
					ExpiresAt: jwt.At(now.Add(time.Hour)),
					IssuedAt:  jwt.At(now),
				},
			}

			if diff := cmp.Diff(want, claims); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the expected claims. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the expected claims.", tests.Success, testID)
		}
	}
}
