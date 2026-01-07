package product

type Product struct {
	ProductID     int64  `db:"product_id"`
	ProductName   string `db:"product_name"`
	ProductGUID   string `db:"product_guid"`
	LatestVersion string `db:"latest_version"`
	DownloadURL   string `db:"download_url"`
}
