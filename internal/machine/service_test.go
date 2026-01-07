package machine_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/registration"
	"winsbygroup.com/regserver/internal/testutil"
)

func TestMachine_GetOrCreate_Create(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	machSvc := machine.NewService(db)

	// Create customer
	c, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Acme",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	tx := db.MustBeginTx(ctx, nil)

	machineID, err := machSvc.GetOrCreate(ctx, tx, c.CustomerID, "MACHINE-123", "John")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}

	if machineID == 0 {
		t.Fatalf("expected non-zero machineID")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify in DB
	var m machine.Machine
	err = db.GetContext(ctx, &m, `
        SELECT machine_id, customer_id, machine_code, user_name
        FROM machine WHERE machine_id = ?
    `, machineID)
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	if m.MachineCode != "MACHINE-123" {
		t.Errorf("expected MACHINE-123, got %s", m.MachineCode)
	}
	if m.UserName != "John" {
		t.Errorf("expected John, got %s", m.UserName)
	}
}

func TestMachine_GetOrCreate_ReusesExisting(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	machSvc := machine.NewService(db)

	c, _ := custSvc.Create(ctx, &customer.Customer{CustomerName: "Acme"})

	// First activation
	tx1 := db.MustBeginTx(ctx, nil)
	id1, err := machSvc.GetOrCreate(ctx, tx1, c.CustomerID, "MACHINE-123", "John")
	if err != nil {
		t.Fatalf("first GetOrCreate: %v", err)
	}
	tx1.Commit()

	// Second activation with same machine_code
	tx2 := db.MustBeginTx(ctx, nil)
	id2, err := machSvc.GetOrCreate(ctx, tx2, c.CustomerID, "MACHINE-123", "John")
	if err != nil {
		t.Fatalf("second GetOrCreate: %v", err)
	}
	tx2.Commit()

	if id1 != id2 {
		t.Fatalf("expected same machineID, got %d and %d", id1, id2)
	}
}

func TestMachine_GetOrCreate_UpdatesUserName(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	machSvc := machine.NewService(db)

	c, _ := custSvc.Create(ctx, &customer.Customer{CustomerName: "Acme"})

	// Create machine
	tx1 := db.MustBeginTx(ctx, nil)
	id, _ := machSvc.GetOrCreate(ctx, tx1, c.CustomerID, "MACHINE-123", "John")
	tx1.Commit()

	// Re-activate with new username
	tx2 := db.MustBeginTx(ctx, nil)
	_, err := machSvc.GetOrCreate(ctx, tx2, c.CustomerID, "MACHINE-123", "Jane")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	tx2.Commit()

	// Verify username updated
	var m machine.Machine
	err = db.GetContext(ctx, &m, `
        SELECT user_name FROM machine WHERE machine_id = ?
    `, id)
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	if m.UserName != "Jane" {
		t.Errorf("expected Jane, got %s", m.UserName)
	}
}

func TestMachine_GetForLicense(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	custSvc := customer.NewService(db)
	prodSvc := product.NewService(db)
	machSvc := machine.NewService(db)
	regSvc := registration.NewService(db)

	// Create customer
	c, err := custSvc.Create(ctx, &customer.Customer{
		CustomerName: "Acme",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	// Create product
	p, err := prodSvc.Create(ctx, &product.Product{
		ProductName: "Widget",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	// Create machine (requires transaction)
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	machineID, err := machSvc.GetOrCreate(ctx, tx, c.CustomerID, "MACHINE-123", "John")
	if err != nil {
		tx.Rollback()
		t.Fatalf("GetOrCreate: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	// Create registration (service manages its own transaction)
	reg := &registration.Registration{
		MachineID:             machineID,
		ProductID:             p.ProductID,
		ExpirationDate:        "2030-01-01", // string, not time.Time
		RegistrationHash:      "dummy-hash",
		FirstRegistrationDate: "2024-01-01",
		LastRegistrationDate:  "2024-01-01",
	}

	_, err = regSvc.Create(ctx, reg)
	if err != nil {
		t.Fatalf("create registration: %v", err)
	}

	// Now exercise the machine relationship query
	machines, err := machSvc.GetForLicense(ctx, c.CustomerID, p.ProductID)
	if err != nil {
		t.Fatalf("GetForLicense: %v", err)
	}

	if len(machines) != 1 {
		t.Fatalf("expected 1 machine, got %d", len(machines))
	}

	if machines[0].MachineCode != "MACHINE-123" {
		t.Errorf("expected MACHINE-123, got %s", machines[0].MachineCode)
	}

	if machines[0].CustomerID != c.CustomerID {
		t.Errorf("expected customerID %d, got %d", c.CustomerID, machines[0].CustomerID)
	}
}
