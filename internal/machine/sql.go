package machine

const getMachineSQL = `
SELECT machine_id, customer_id, machine_code, user_name
FROM machine
WHERE customer_id = ? AND machine_code = ?
`

const getMachineByIDSQL = `
SELECT machine_id, customer_id, machine_code, user_name
FROM machine
WHERE machine_id = ?
`

const getActiveForLicenseSQL = `
SELECT m.machine_id, m.customer_id, m.machine_code, m.user_name
FROM machine m
JOIN registration r ON r.machine_id = m.machine_id
WHERE m.customer_id = ?
  AND r.product_id = ?
  AND r.expiration_date >= DATE('now')
ORDER BY m.machine_code
`

const getForLicenseSQL = `
SELECT m.machine_id, m.customer_id, m.machine_code, m.user_name
FROM machine m
JOIN registration r ON r.machine_id = m.machine_id
WHERE m.customer_id = ? AND r.product_id = ?
ORDER BY m.machine_code
`

const createMachineSQL = `
INSERT INTO machine (customer_id, machine_code, user_name)
VALUES (?, ?, ?)
`

const updateUserNameSQL = `
UPDATE machine
SET user_name = ?
WHERE machine_id = ?
`
