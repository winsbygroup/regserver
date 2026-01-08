/*
  sqlite DDL to create the registrations database
*/

CREATE TABLE IF NOT EXISTS customer (
    customer_id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_name VARCHAR(255) NOT NULL UNIQUE COLLATE NOCASE,
    contact_name VARCHAR(255),
    phone VARCHAR(255),
    email VARCHAR(255),
    notes TEXT
);

CREATE TABLE IF NOT EXISTS machine (
    machine_id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    machine_code VARCHAR(255) NOT NULL,
    user_name VARCHAR(255),
    FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_machine_customer_id ON machine (customer_id ASC);


CREATE TABLE IF NOT EXISTS product (
    product_id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_name VARCHAR(255) NOT NULL UNIQUE COLLATE NOCASE,
    product_guid VARCHAR(36) NOT NULL UNIQUE COLLATE NOCASE,
	latest_version VARCHAR(10) NOT NULL,
	download_url VARCHAR(255) NOT NULL DEFAULT ('https://download.example.com/{product}')
);

CREATE TABLE IF NOT EXISTS registration (
    machine_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    expiration_date VARCHAR(10) NOT NULL,
    registration_hash CHAR(28) NOT NULL,
    first_registration_date VARCHAR(10),
    last_registration_date VARCHAR(10),
    installed_version VARCHAR(20) NOT NULL DEFAULT '',
    CONSTRAINT pk_registration PRIMARY KEY (machine_id, product_id),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE,
    FOREIGN KEY (machine_id) REFERENCES machine (machine_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_registration_machine_id ON registration (machine_id ASC);
CREATE INDEX IF NOT EXISTS idx_registration_product_id ON registration (product_id ASC);


CREATE TABLE IF NOT EXISTS license (
    customer_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    license_key VARCHAR(36) NOT NULL COLLATE NOCASE,
    license_count INTEGER NOT NULL,
    is_subscription INTEGER NOT NULL,
    license_term INTEGER NOT NULL,
    start_date VARCHAR(10),
    expiration_date VARCHAR(10),
    maint_expiration_date VARCHAR(10) NOT NULL DEFAULT '9999-12-31',
    max_product_version VARCHAR(255),
    CONSTRAINT pk_license PRIMARY KEY (customer_id, product_id),
    FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_license_customer_id ON license (customer_id ASC);
CREATE INDEX IF NOT EXISTS idx_license_product_id ON license (product_id ASC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_license_key ON license (license_key);


CREATE TABLE IF NOT EXISTS feature (
    feature_id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    feature_name VARCHAR(255) NOT NULL,	
    feature_type INTEGER NOT NULL CHECK (feature_type in (0,1,2)) DEFAULT 0,		
    allowed_values VARCHAR(255),
    default_value VARCHAR(255),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_feature_product_id ON feature (product_id ASC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feature_product_name ON feature (product_id, feature_name COLLATE NOCASE);


CREATE TABLE IF NOT EXISTS license_feature (
    customer_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    feature_value VARCHAR(255) NOT NULL,
    CONSTRAINT pk_license_feature PRIMARY KEY (customer_id, product_id, feature_id),
    FOREIGN KEY (feature_id) REFERENCES feature (feature_id),
    FOREIGN KEY (customer_id, product_id) REFERENCES license (customer_id, product_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_license_feature_feature_id ON license_feature (feature_id ASC);
CREATE INDEX IF NOT EXISTS idx_license_feature_custid_prodid ON license_feature (customer_id ASC, product_id ASC);
