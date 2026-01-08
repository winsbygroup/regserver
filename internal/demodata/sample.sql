-- Demo data for regserver
-- This data is loaded when running with -demo flag on a new database

-- Customers
INSERT INTO customer (customer_id, customer_name, contact_name, phone, email, notes) VALUES
(1, 'Acme Corporation', 'John Smith', '555-0100', 'john.smith@acme.example.com', 'Enterprise customer since 2020'),
(2, 'TechStart Inc', 'Sarah Johnson', '555-0200', 'sarah@techstart.example.com', 'Startup - 10 seat license'),
(3, 'Global Industries', 'Mike Chen', '555-0300', 'mchen@global.example.com', 'Multi-product customer');

-- Products
INSERT INTO product (product_id, product_name, product_guid, latest_version, download_url) VALUES
(1, 'DataMapper Pro', 'a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d', '3.2.1', 'https://example.com/downloads/datamapper-3.2.1.zip'),
(2, 'ReportBuilder', 'f6e5d4c3-b2a1-4f5e-9d8c-7b6a5f4e3d2c', '2.0.0', 'https://example.com/downloads/reportbuilder-2.0.0.zip');

-- Licenses
-- Acme: DataMapper Pro - 25 seats, perpetual with maintenance
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(1, 1, '11111111-1111-1111-1111-111111111111', 25, 0, 0, '2024-01-01', '9999-12-31', '2025-12-31', '');

-- Acme: ReportBuilder - 10 seats, perpetual, no maintenance
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(1, 2, '22222222-2222-2222-2222-222222222222', 10, 0, 0, '2024-01-01', '9999-12-31', '9999-12-31', '');

-- TechStart: DataMapper Pro - 10 seats, annual subscription
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(2, 1, '33333333-3333-3333-3333-333333333333', 10, 1, 12, '2024-06-01', '2025-06-01', '2025-06-01', '');

-- Global Industries: DataMapper Pro - 50 seats, perpetual, version-locked
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(3, 1, '44444444-4444-4444-4444-444444444444', 50, 0, 0, '2023-01-01', '9999-12-31', '2024-12-31', '2.5.0');

-- Global Industries: ReportBuilder - 5 seats, monthly subscription
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(3, 2, '55555555-5555-5555-5555-555555555555', 5, 1, 1, '2024-12-01', '2025-01-01', '2025-01-01', '');

-- Features for DataMapper Pro
INSERT INTO feature (feature_id, product_id, feature_name, feature_type, allowed_values, default_value) VALUES
(1, 1, 'MaxRecords', 0, '', '10000'),
(2, 1, 'ExportFormats', 2, 'CSV|JSON|XML|Excel', 'CSV|JSON'),
(3, 1, 'CloudSync', 2, 'true|false', 'false'),
(4, 1, 'SupportTier', 1, '', 'Standard');

-- Features for ReportBuilder
INSERT INTO feature (feature_id, product_id, feature_name, feature_type, allowed_values, default_value) VALUES
(5, 2, 'MaxReports', 0, '', '50'),
(6, 2, 'ScheduledReports', 2, 'true|false', 'false'),
(7, 2, 'OutputFormats', 2, 'PDF|HTML|Excel', 'PDF');

-- Feature value overrides (customer-specific)
-- Acme gets enterprise features on DataMapper
INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
(1, 1, 1, '999999'),
(1, 1, 2, 'CSV|JSON|XML|Excel'),
(1, 1, 3, 'true'),
(1, 1, 4, 'Enterprise');

-- TechStart gets startup tier on DataMapper
INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
(2, 1, 1, '50000'),
(2, 1, 4, 'Startup');

-- Global Industries gets custom ReportBuilder limits
INSERT INTO license_feature (customer_id, product_id, feature_id, feature_value) VALUES
(3, 2, 5, '200'),
(3, 2, 6, 'true');

-- Machines
-- Acme Corporation machines
INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
(1, 1, 'ACME-WS-001', 'jsmith'),
(2, 1, 'ACME-WS-002', 'bjones'),
(3, 1, 'ACME-WS-003', 'mwilson'),
(4, 1, 'ACME-LAPTOP-001', 'jsmith'),
(5, 1, 'ACME-SERVER-01', 'admin');

-- TechStart machines
INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
(6, 2, 'TS-DEV-01', 'sarah'),
(7, 2, 'TS-DEV-02', 'alex'),
(8, 2, 'TS-DEV-03', 'chris');

-- Global Industries machines
INSERT INTO machine (machine_id, customer_id, machine_code, user_name) VALUES
(9, 3, 'GI-NYC-WS001', 'mchen'),
(10, 3, 'GI-NYC-WS002', 'lpark'),
(11, 3, 'GI-LA-WS001', 'dkim'),
(12, 3, 'GI-CHI-WS001', 'rjohnson');

-- Registrations (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version)
-- Note: registration_hash is a placeholder for demo purposes

-- Acme DataMapper Pro registrations (5 of 25 seats used)
INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version) VALUES
(1, 1, '2025-12-31', 'demo-hash-acme-dm-001', '2024-01-15', '2024-12-01', '3.2.1'),
(2, 1, '2025-12-31', 'demo-hash-acme-dm-002', '2024-02-01', '2024-11-15', '3.2.0'),
(3, 1, '2025-12-31', 'demo-hash-acme-dm-003', '2024-03-10', '2024-12-05', '3.2.1'),
(4, 1, '2025-12-31', 'demo-hash-acme-dm-004', '2024-06-01', '2024-12-01', '3.1.0'),
(5, 1, '2025-12-31', 'demo-hash-acme-dm-005', '2024-01-15', '2024-10-20', '3.2.1');

-- Acme ReportBuilder registrations (3 of 10 seats used)
INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version) VALUES
(1, 2, '9999-12-31', 'demo-hash-acme-rb-001', '2024-02-01', '2024-11-01', '2.0.0'),
(2, 2, '9999-12-31', 'demo-hash-acme-rb-002', '2024-02-15', '2024-10-15', '2.0.0'),
(5, 2, '9999-12-31', 'demo-hash-acme-rb-003', '2024-03-01', '2024-12-01', '2.0.0');

-- TechStart DataMapper Pro registrations (3 of 10 seats used)
INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version) VALUES
(6, 1, '2025-06-01', 'demo-hash-ts-dm-001', '2024-06-15', '2024-12-01', '3.2.1'),
(7, 1, '2025-06-01', 'demo-hash-ts-dm-002', '2024-07-01', '2024-11-20', '3.2.1'),
(8, 1, '2025-06-01', 'demo-hash-ts-dm-003', '2024-08-01', '2024-12-05', '3.2.0');

-- Global Industries DataMapper Pro registrations (4 of 50 seats used, version-locked to 2.5)
INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version) VALUES
(9, 1, '2024-12-31', 'demo-hash-gi-dm-001', '2023-01-20', '2024-06-15', '2.5.0'),
(10, 1, '2024-12-31', 'demo-hash-gi-dm-002', '2023-02-01', '2024-07-01', '2.5.0'),
(11, 1, '2024-12-31', 'demo-hash-gi-dm-003', '2023-03-15', '2024-08-20', '2.4.0'),
(12, 1, '2024-12-31', 'demo-hash-gi-dm-004', '2023-04-01', '2024-05-10', '2.5.0');

-- Global Industries ReportBuilder registrations (2 of 5 seats used)
INSERT INTO registration (machine_id, product_id, expiration_date, registration_hash, first_registration_date, last_registration_date, installed_version) VALUES
(9, 2, '2025-01-01', 'demo-hash-gi-rb-001', '2024-12-01', '2024-12-15', '2.0.0'),
(10, 2, '2025-01-01', 'demo-hash-gi-rb-002', '2024-12-05', '2024-12-10', '2.0.0');
