package registration

const getRegistrationSQL = `
SELECT
    machine_id,
    product_id,
    expiration_date,
    registration_hash,
    first_registration_date,
    last_registration_date,
    installed_version
FROM registration
WHERE machine_id = ? AND product_id = ?
`

const getRegistrationsForMachineSQL = `
SELECT
    machine_id,
    product_id,
    expiration_date,
    registration_hash,
    first_registration_date,
    last_registration_date,
    installed_version
FROM registration
WHERE machine_id = ?
ORDER BY product_id
`

const createRegistrationSQL = `
INSERT INTO registration (
    machine_id,
    product_id,
    expiration_date,
    registration_hash,
    first_registration_date,
    last_registration_date
) VALUES (?, ?, ?, ?, ?, ?)
`

const updateRegistrationSQL = `
UPDATE registration
SET
    expiration_date = ?,
    registration_hash = ?,
    first_registration_date = ?,
    last_registration_date = ?
WHERE machine_id = ? AND product_id = ?
`

/*
- first_registration_date is only set on insert
- last_registration_date is always updated
- expiration_date is refreshed from customer_product
- registration_hash is updated (your original code used machineCode)
*/
const upsertRegistrationSQL = `
INSERT INTO registration (
    machine_id,
    product_id,
    expiration_date,
    registration_hash,
    first_registration_date,
    last_registration_date
) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(machine_id, product_id) DO UPDATE SET
    expiration_date = excluded.expiration_date,
    registration_hash = excluded.registration_hash,
    last_registration_date = excluded.last_registration_date
`

const deleteRegistrationSQL = `
DELETE FROM registration
WHERE machine_id = ? AND product_id = ?
`

const updateInstalledVersionSQL = `
UPDATE registration
SET installed_version = ?
WHERE machine_id = ? AND product_id = ?
`
