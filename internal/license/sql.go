package license

const getLicenseSQL = `
SELECT
    customer_id,
    product_id,
    license_key,
    license_count,
    is_subscription,
    license_term,
    start_date,
    expiration_date,
    maint_expiration_date,
    max_product_version
FROM license
WHERE customer_id = ? AND product_id = ?
`

const getLicensesSQL = `
SELECT
    customer_id,
    product_id,
    license_key,
    license_count,
    is_subscription,
    license_term,
    start_date,
    expiration_date,
    maint_expiration_date,
    max_product_version
FROM license
WHERE customer_id = ?
ORDER BY product_id
`

const getUnlicensedProductsSQL = `
SELECT p.product_id, p.product_name
FROM product p
WHERE p.product_id NOT IN (
    SELECT product_id
    FROM license
    WHERE customer_id = ?
)
ORDER BY p.product_name
`

const createLicenseSQL = `
INSERT INTO license (
    customer_id,
    product_id,
    license_key,
    license_count,
    is_subscription,
    license_term,
    start_date,
    expiration_date,
    maint_expiration_date,
    max_product_version
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

const updateLicenseSQL = `
UPDATE license
SET
    license_count = ?,
    is_subscription = ?,
    license_term = ?,
    start_date = ?,
    expiration_date = ?,
    maint_expiration_date = ?,
    max_product_version = ?
WHERE customer_id = ? AND product_id = ?
`

const deleteLicenseSQL = `
DELETE FROM license
WHERE customer_id = ? AND product_id = ?
`

const getExpiredLicensesSQL = `
SELECT
    c.customer_name,
    c.contact_name,
    c.email,
    p.product_name,
    l.expiration_date,
    l.maint_expiration_date
FROM license l
JOIN customer c ON c.customer_id = l.customer_id
JOIN product p ON p.product_id = l.product_id
WHERE l.expiration_date < ? OR l.maint_expiration_date < ?
ORDER BY l.expiration_date DESC
`

const getLicenseByKeySQL = `
SELECT
    customer_id,
    product_id,
    license_key,
    license_count,
    is_subscription,
    license_term,
    start_date,
    expiration_date,
    maint_expiration_date,
    max_product_version
FROM license
WHERE license_key = ?
`
