package product

const getAllProductsSQL = `
SELECT product_id, product_name, product_guid, latest_version, download_url
FROM product
ORDER BY product_name
`

const getProductSQL = `
SELECT product_id, product_name, product_guid, latest_version, download_url
FROM product
WHERE product_id = ?
`

const getProductByGUIDSQL = `
SELECT product_id, product_name, product_guid, latest_version, download_url
FROM product
WHERE product_guid = ?
`

const createProductSQL = `
INSERT INTO product (
    product_name, product_guid, latest_version, download_url
) VALUES (?, ?, ?, ?)
`

const updateProductSQL = `
UPDATE product
SET product_name = ?, product_guid = ?, latest_version = ?, download_url = ?
WHERE product_id = ?
`

const deleteProductSQL = `
DELETE FROM product
WHERE product_id = ?
`
