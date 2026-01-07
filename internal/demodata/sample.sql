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
(1, 1, '11111111-1111-1111-1111-111111111111', 25, 0, 0, '2024-01-01', '9998-12-31', '2025-12-31', '');

-- Acme: ReportBuilder - 10 seats, perpetual, no maintenance
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(1, 2, '22222222-2222-2222-2222-222222222222', 10, 0, 0, '2024-01-01', '9998-12-31', '9998-12-31', '');

-- TechStart: DataMapper Pro - 10 seats, annual subscription
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(2, 1, '33333333-3333-3333-3333-333333333333', 10, 1, 12, '2024-06-01', '2025-06-01', '2025-06-01', '');

-- Global Industries: DataMapper Pro - 50 seats, perpetual, version-locked
INSERT INTO license (customer_id, product_id, license_key, license_count, is_subscription, license_term, start_date, expiration_date, maint_expiration_date, max_product_version) VALUES
(3, 1, '44444444-4444-4444-4444-444444444444', 50, 0, 0, '2023-01-01', '9998-12-31', '2024-12-31', '2.5');

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
