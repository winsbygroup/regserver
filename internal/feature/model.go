package feature

type Feature struct {
	FeatureID     int64  `db:"feature_id"`
	ProductID     int64  `db:"product_id"`
	FeatureName   string `db:"feature_name"`
	FeatureType   int    `db:"feature_type"`
	AllowedValues string `db:"allowed_values"`
	DefaultValue  string `db:"default_value"`
}
