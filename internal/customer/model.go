package customer

type Customer struct {
	CustomerID   int64  `db:"customer_id"`
	CustomerName string `db:"customer_name"`
	ContactName  string `db:"contact_name"`
	Phone        string `db:"phone"`
	Email        string `db:"email"`
	Notes        string `db:"notes"`
}
