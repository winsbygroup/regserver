package featurevalue

type FeatureValue struct {
	CustomerID   int64  `db:"customer_id"`
	ProductID    int64  `db:"product_id"`
	FeatureID    int64  `db:"feature_id"`
	FeatureValue string `db:"feature_value"`
}
