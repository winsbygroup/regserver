package customer_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/sqlite"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestCustomerLifecycle(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := customer.NewService(db)

	// Create
	c := &customer.Customer{
		CustomerName: "Acme Corp",
		ContactName:  "John Doe",
		Phone:        "555-1234",
		Email:        "john@example.com",
		Notes:        "VIP customer",
	}

	created, err := svc.Create(ctx, c)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get
	got, err := svc.Get(ctx, created.CustomerID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.CustomerName != "Acme Corp" {
		t.Errorf("expected name %q, got %q", "Acme Corp", got.CustomerName)
	}

	// Update
	got.ContactName = "Jane Smith"
	if err := svc.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := svc.Get(ctx, got.CustomerID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}

	if updated.ContactName != "Jane Smith" {
		t.Errorf("expected updated contact %q, got %q", "Jane Smith", updated.ContactName)
	}

	// Delete
	if err := svc.Delete(ctx, updated.CustomerID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = svc.Get(ctx, updated.CustomerID)
	if err == nil {
		t.Fatalf("expected error getting deleted customer")
	}
}

func TestCustomerDuplicateNameCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	svc := customer.NewService(db)

	// Create first customer
	_, err := svc.Create(ctx, &customer.Customer{
		CustomerName: "Acme Corp",
		ContactName:  "John Doe",
	})
	if err != nil {
		t.Fatalf("create first customer: %v", err)
	}

	// Try to create customer with same name but different case
	testCases := []string{"ACME CORP", "acme corp", "AcMe CoRp"}
	for _, name := range testCases {
		_, err = svc.Create(ctx, &customer.Customer{
			CustomerName: name,
			ContactName:  "Jane Smith",
		})
		if err == nil {
			t.Errorf("expected error for duplicate name %q, got none", name)
		}
		if !sqlite.IsUniqueConstraintError(err) {
			t.Errorf("expected unique constraint error for %q, got: %v", name, err)
		}
	}
}
