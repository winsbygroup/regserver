package featurevalue

const getFeatureValuesSQL = `
SELECT
    customer_id,
    product_id,
    feature_id,
    feature_value
FROM license_feature
WHERE customer_id = ? AND product_id = ?
ORDER BY feature_id
`

const updateFeatureValueSQL = `
INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value)
VALUES (?, ?, ?, ?)
ON CONFLICT (customer_id, product_id, feature_id) DO UPDATE SET feature_value = excluded.feature_value
`
