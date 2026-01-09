package sqlite_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"winsbygroup.com/regserver/internal/customer"
	"winsbygroup.com/regserver/internal/feature"
	"winsbygroup.com/regserver/internal/license"
	"winsbygroup.com/regserver/internal/machine"
	"winsbygroup.com/regserver/internal/product"
	"winsbygroup.com/regserver/internal/testutil"
)

// countRows returns the number of rows in a table
func countRows(t *testing.T, db *sqlx.DB, table string) int {
	t.Helper()
	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM "+table); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

// countWhere returns the count from a query with args
func countWhere(t *testing.T, db *sqlx.DB, query string, args ...interface{}) int {
	t.Helper()
	var count int
	if err := db.Get(&count, query, args...); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	return count
}

// insertTestData executes SQL statements to set up test data
func insertTestData(t *testing.T, db *sqlx.DB, sql string) {
	t.Helper()
	if _, err := db.Exec(sql); err != nil {
		t.Fatalf("insert test data: %v", err)
	}
}

// TestCascadeDeleteCustomer verifies that deleting a customer cascades to:
// - machines (direct FK)
// - licenses (direct FK)
// - registrations (via machine FK)
// - license_feature (via license FK)
func TestCascadeDeleteCustomer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Create test data: 2 customers, 1 product
	// Customer 1 has: 2 machines, 1 license, 2 registrations, 1 feature value
	// Customer 2 has: 1 machine (to verify it's not deleted)
	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One'),
			(2, 'Customer Two');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com');

		INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
			(1, 1, 'MACHINE-1A', 'user1'),
			(2, 1, 'MACHINE-1B', 'user2'),
			(3, 2, 'MACHINE-2A', 'user3');

		INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, maint_expiration_date) VALUES
			(1, 1, 'LIC-001', 5, 0, 0, '9999-12-31');

		INSERT INTO feature (feature_id, product_id, feature_name, feature_type, default_value) VALUES
			(1, 1, 'MaxUsers', 0, '10');

		INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
			(1, 1, 1, '100');

		INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date) VALUES
			(1, 1, '2030-01-01', 'hash1', '2024-01-01', '2024-01-01'),
			(2, 1, '2030-01-01', 'hash2', '2024-01-01', '2024-01-01');
	`)

	// Verify initial state
	if got := countWhere(t, db, "SELECT COUNT(*) FROM machine WHERE customer_id = 1"); got != 2 {
		t.Fatalf("expected 2 machines for customer 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license WHERE customer_id = 1"); got != 1 {
		t.Fatalf("expected 1 license for customer 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE customer_id = 1"); got != 1 {
		t.Fatalf("expected 1 feature value for customer 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE machine_id IN (1, 2)"); got != 2 {
		t.Fatalf("expected 2 registrations for customer 1's machines, got %d", got)
	}

	// Delete customer 1
	custSvc := customer.NewService(db)
	if err := custSvc.Delete(ctx, 1); err != nil {
		t.Fatalf("delete customer: %v", err)
	}

	// Verify cascade deletions
	if got := countWhere(t, db, "SELECT COUNT(*) FROM machine WHERE customer_id = 1"); got != 0 {
		t.Errorf("expected 0 machines after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license WHERE customer_id = 1"); got != 0 {
		t.Errorf("expected 0 licenses after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE customer_id = 1"); got != 0 {
		t.Errorf("expected 0 feature values after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE machine_id IN (1, 2)"); got != 0 {
		t.Errorf("expected 0 registrations after delete, got %d", got)
	}

	// Verify customer 2's data is intact
	if got := countWhere(t, db, "SELECT COUNT(*) FROM machine WHERE customer_id = 2"); got != 1 {
		t.Errorf("expected customer 2's machine to remain, got %d", got)
	}
}

// TestCascadeDeleteProduct verifies that deleting a product cascades to:
// - features (direct FK)
// - licenses (direct FK)
// - registrations (direct FK)
// - license_feature (via license FK)
func TestCascadeDeleteProduct(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Create test data: 1 customer, 2 products
	// Product 1 has: 2 features, 1 license, 2 registrations, 1 feature value
	// Product 2 has: 1 feature (to verify it's not deleted)
	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com/1'),
			(2, 'Product Two', 'guid-2', '1.0', 'http://example.com/2');

		INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
			(1, 1, 'MACHINE-1', 'user1'),
			(2, 1, 'MACHINE-2', 'user2');

		INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, maint_expiration_date) VALUES
			(1, 1, 'LIC-001', 5, 0, 0, '9999-12-31');

		INSERT INTO feature (feature_id, product_id, feature_name, feature_type, default_value) VALUES
			(1, 1, 'Feature1A', 0, '10'),
			(2, 1, 'Feature1B', 0, '20'),
			(3, 2, 'Feature2A', 0, '30');

		INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
			(1, 1, 1, '100');

		INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date) VALUES
			(1, 1, '2030-01-01', 'hash1', '2024-01-01', '2024-01-01'),
			(2, 1, '2030-01-01', 'hash2', '2024-01-01', '2024-01-01');
	`)

	// Verify initial state
	if got := countWhere(t, db, "SELECT COUNT(*) FROM feature WHERE product_id = 1"); got != 2 {
		t.Fatalf("expected 2 features for product 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license WHERE product_id = 1"); got != 1 {
		t.Fatalf("expected 1 license for product 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE product_id = 1"); got != 2 {
		t.Fatalf("expected 2 registrations for product 1, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE product_id = 1"); got != 1 {
		t.Fatalf("expected 1 feature value for product 1, got %d", got)
	}

	// Delete product 1
	prodSvc := product.NewService(db)
	if err := prodSvc.Delete(ctx, 1); err != nil {
		t.Fatalf("delete product: %v", err)
	}

	// Verify cascade deletions
	if got := countWhere(t, db, "SELECT COUNT(*) FROM feature WHERE product_id = 1"); got != 0 {
		t.Errorf("expected 0 features after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license WHERE product_id = 1"); got != 0 {
		t.Errorf("expected 0 licenses after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE product_id = 1"); got != 0 {
		t.Errorf("expected 0 registrations after delete, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE product_id = 1"); got != 0 {
		t.Errorf("expected 0 feature values after delete, got %d", got)
	}

	// Verify product 2's data is intact
	if got := countWhere(t, db, "SELECT COUNT(*) FROM feature WHERE product_id = 2"); got != 1 {
		t.Errorf("expected product 2's feature to remain, got %d", got)
	}
}

// TestCascadeDeleteMachine verifies that deleting a machine cascades to:
// - registrations (direct FK)
func TestCascadeDeleteMachine(t *testing.T) {
	db := testutil.NewTestDB(t)

	// Create test data: 1 customer, 1 product, 2 machines
	// Machine 1 has 2 registrations, Machine 2 has 1 registration
	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com/1'),
			(2, 'Product Two', 'guid-2', '1.0', 'http://example.com/2');

		INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
			(1, 1, 'MACHINE-1', 'user1'),
			(2, 1, 'MACHINE-2', 'user2');

		INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date) VALUES
			(1, 1, '2030-01-01', 'hash1a', '2024-01-01', '2024-01-01'),
			(1, 2, '2030-01-01', 'hash1b', '2024-01-01', '2024-01-01'),
			(2, 1, '2030-01-01', 'hash2', '2024-01-01', '2024-01-01');
	`)

	// Verify initial state
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE machine_id = 1"); got != 2 {
		t.Fatalf("expected 2 registrations for machine 1, got %d", got)
	}
	totalBefore := countRows(t, db, "registration")

	// Delete machine 1 directly (no service method exists)
	if _, err := db.Exec("DELETE FROM machine WHERE machine_id = 1"); err != nil {
		t.Fatalf("delete machine: %v", err)
	}

	// Verify cascade deletion
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE machine_id = 1"); got != 0 {
		t.Errorf("expected 0 registrations after delete, got %d", got)
	}

	// Verify machine 2's registration is intact
	totalAfter := countRows(t, db, "registration")
	if totalAfter != totalBefore-2 {
		t.Errorf("expected %d registrations remaining, got %d", totalBefore-2, totalAfter)
	}
}

// TestCascadeDeleteLicense verifies that deleting a license cascades to:
// - license_feature (direct FK on composite key)
func TestCascadeDeleteLicense(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Create test data: 2 customers, 1 product, each with a license and feature values
	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One'),
			(2, 'Customer Two');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com');

		INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, maint_expiration_date) VALUES
			(1, 1, 'LIC-001', 5, 0, 0, '9999-12-31'),
			(2, 1, 'LIC-002', 5, 0, 0, '9999-12-31');

		INSERT INTO feature (feature_id, product_id, feature_name, feature_type, default_value) VALUES
			(1, 1, 'MaxUsers', 0, '10'),
			(2, 1, 'MaxSessions', 0, '5');

		INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
			(1, 1, 1, '100'),
			(1, 1, 2, '50'),
			(2, 1, 1, '200');
	`)

	// Verify initial state
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE customer_id = 1 AND product_id = 1"); got != 2 {
		t.Fatalf("expected 2 feature values for license (1,1), got %d", got)
	}

	// Delete license for customer 1
	licSvc := license.NewService(db)
	if err := licSvc.Delete(ctx, 1, 1); err != nil {
		t.Fatalf("delete license: %v", err)
	}

	// Verify cascade deletion
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE customer_id = 1 AND product_id = 1"); got != 0 {
		t.Errorf("expected 0 feature values after delete, got %d", got)
	}

	// Verify customer 2's feature values are intact
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license_feature WHERE customer_id = 2"); got != 1 {
		t.Errorf("expected customer 2's feature value to remain, got %d", got)
	}
}

// TestCascadeDeleteFeature verifies that deleting a feature does NOT cascade
// to license_feature (no ON DELETE CASCADE on feature_id FK)
func TestCascadeDeleteFeature(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Create test data with a license_feature referencing a feature
	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com');

		INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, maint_expiration_date) VALUES
			(1, 1, 'LIC-001', 5, 0, 0, '9999-12-31');

		INSERT INTO feature (feature_id, product_id, feature_name, feature_type, default_value) VALUES
			(1, 1, 'MaxUsers', 0, '10');

		INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
			(1, 1, 1, '100');
	`)

	// Attempting to delete the feature should fail due to FK constraint
	// (license_feature has FK to feature but no ON DELETE CASCADE)
	featSvc := feature.NewService(db)
	err := featSvc.Delete(ctx, 1)
	if err == nil {
		t.Error("expected FK constraint error when deleting feature with dependent license_feature, got nil")
	}

	// Clean up the license_feature first, then feature should delete
	if _, err := db.Exec("DELETE FROM license_feature WHERE feature_id = 1"); err != nil {
		t.Fatalf("cleanup license_feature: %v", err)
	}

	if err := featSvc.Delete(ctx, 1); err != nil {
		t.Errorf("expected feature delete to succeed after cleanup, got: %v", err)
	}
}

// TestCascadeMultiLevel verifies multi-level cascade:
// customer -> machine -> registration (two-level cascade)
func TestCascadeMultiLevel(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com');

		INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
			(1, 1, 'MACHINE-1', 'user1'),
			(2, 1, 'MACHINE-2', 'user2');

		INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date) VALUES
			(1, 1, '2030-01-01', 'hash1', '2024-01-01', '2024-01-01'),
			(2, 1, '2030-01-01', 'hash2', '2024-01-01', '2024-01-01');
	`)

	// Verify we have registrations
	if got := countRows(t, db, "registration"); got != 2 {
		t.Fatalf("expected 2 registrations, got %d", got)
	}

	// Delete customer - should cascade through machines to registrations
	custSvc := customer.NewService(db)
	if err := custSvc.Delete(ctx, 1); err != nil {
		t.Fatalf("delete customer: %v", err)
	}

	// All registrations should be gone (via machine cascade)
	if got := countRows(t, db, "registration"); got != 0 {
		t.Errorf("expected 0 registrations after customer delete, got %d", got)
	}
	if got := countRows(t, db, "machine"); got != 0 {
		t.Errorf("expected 0 machines after customer delete, got %d", got)
	}
}

// TestForeignKeyEnforcementEnabled verifies that FK constraints are enforced
func TestForeignKeyEnforcementEnabled(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	machSvc := machine.NewService(db)

	// Try to create a machine for a non-existent customer
	tx := db.MustBeginTx(ctx, nil)
	_, err := machSvc.GetOrCreate(ctx, tx, 99999, "FAKE-MACHINE", "user")
	if err == nil {
		tx.Rollback()
		t.Fatal("expected FK error when creating machine for non-existent customer")
	}
	tx.Rollback()
}

// TestCascadeDoesNotAffectUnrelatedData verifies isolation between entities
func TestCascadeDoesNotAffectUnrelatedData(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	insertTestData(t, db, `
		INSERT INTO customer (customer_id, customer_name) VALUES
			(1, 'Customer One'),
			(2, 'Customer Two');

		INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
			(1, 'Product One', 'guid-1', '1.0', 'http://example.com'),
			(2, 'Product Two', 'guid-2', '1.0', 'http://example.com');

		INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
			(1, 1, 'MACHINE-1', 'user1'),
			(2, 2, 'MACHINE-2', 'user2');

		INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, maint_expiration_date) VALUES
			(1, 1, 'LIC-1-1', 5, 0, 0, '9999-12-31'),
			(2, 2, 'LIC-2-2', 5, 0, 0, '9999-12-31');

		INSERT INTO feature (feature_id, product_id, feature_name, feature_type, default_value) VALUES
			(1, 1, 'Feature1', 0, '10'),
			(2, 2, 'Feature2', 0, '20');

		INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date) VALUES
			(1, 1, '2030-01-01', 'hash1', '2024-01-01', '2024-01-01'),
			(2, 2, '2030-01-01', 'hash2', '2024-01-01', '2024-01-01');
	`)

	// Delete customer 1
	custSvc := customer.NewService(db)
	if err := custSvc.Delete(ctx, 1); err != nil {
		t.Fatalf("delete customer 1: %v", err)
	}

	// Customer 2's data should be completely intact
	if got := countRows(t, db, "customer"); got != 1 {
		t.Errorf("expected 1 customer remaining, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM machine WHERE customer_id = 2"); got != 1 {
		t.Errorf("expected customer 2's machine intact, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM license WHERE customer_id = 2"); got != 1 {
		t.Errorf("expected customer 2's license intact, got %d", got)
	}
	if got := countWhere(t, db, "SELECT COUNT(*) FROM registration WHERE machine_id = 2"); got != 1 {
		t.Errorf("expected customer 2's registration intact, got %d", got)
	}

	// Products and features should be intact (they belong to products, not customers)
	if got := countRows(t, db, "product"); got != 2 {
		t.Errorf("expected 2 products remaining, got %d", got)
	}
	if got := countRows(t, db, "feature"); got != 2 {
		t.Errorf("expected 2 features remaining, got %d", got)
	}
}
