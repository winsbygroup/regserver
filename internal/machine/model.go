package machine

type Machine struct {
	MachineID   int64  `db:"machine_id"`
	CustomerID  int64  `db:"customer_id"`
	MachineCode string `db:"machine_code"`
	UserName    string `db:"user_name"`
}
