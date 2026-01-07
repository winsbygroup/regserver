package registration

type Registration struct {
	MachineID             int64  `db:"machine_id"`
	ProductID             int64  `db:"product_id"`
	ExpirationDate        string `db:"expiration_date"`
	RegistrationHash      string `db:"registration_hash"`
	FirstRegistrationDate string `db:"first_registration_date"`
	LastRegistrationDate  string `db:"last_registration_date"`
	InstalledVersion      string `db:"installed_version"`
}
