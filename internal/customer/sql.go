package customer

const getAllCustomersSQL = `
SELECT customer_id, customer_name, contact_name, phone, email, notes
FROM customer
ORDER BY customer_name
`

const getCustomerSQL = `
SELECT customer_id, customer_name, contact_name, phone, email, notes
FROM customer
WHERE customer_id = ?
`

const createCustomerSQL = `
INSERT INTO customer (
    customer_name, contact_name, phone, email, notes
) VALUES (?, ?, ?, ?, ?)
`

const updateCustomerSQL = `
UPDATE customer
SET customer_name = ?, contact_name = ?, phone = ?, email = ?, notes = ?
WHERE customer_id = ?
`

const deleteCustomerSQL = `
DELETE FROM customer
WHERE customer_id = ?
`

const customerExistsSQL = `
SELECT EXISTS(
    SELECT 1 FROM customer WHERE customer_id = ?
)
`
