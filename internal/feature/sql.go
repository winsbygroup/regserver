package feature

const getFeaturesForProductSQL = `
SELECT
    feature_id,
    product_id,
    feature_name,
    feature_type,
    allowed_values,
    default_value
FROM feature
WHERE product_id = ?
ORDER BY feature_name
`

const getFeatureSQL = `
SELECT
    feature_id,
    product_id,
    feature_name,
    feature_type,
    allowed_values,
    default_value
FROM feature
WHERE feature_id = ?
`

const createFeatureSQL = `
INSERT INTO feature (
    product_id,
    feature_name,
    feature_type,
    allowed_values,
    default_value
) VALUES (?, ?, ?, ?, ?)
`

const updateFeatureSQL = `
UPDATE feature
SET
    feature_name = ?,
    feature_type = ?,
    allowed_values = ?,
    default_value = ?
WHERE feature_id = ?
`

const deleteFeatureSQL = `
DELETE FROM feature
WHERE feature_id = ?
`
