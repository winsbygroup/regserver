PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
-- TABLE customer
INSERT INTO customer VALUES(1,'Winsby Group LLC','Doug Winsby','314-432-9222','dougw@winsbygroup.com','Great company!');
-- TABLE machine
INSERT INTO machine VALUES(1,1,'8Er8wNbGzT/7NU+1Wq+b0r9FBfs=nWxB5pHbLwJx/LbewudPWXecK3c=','dell-xps');
-- TABLE product
INSERT INTO product VALUES(1,'AceMapper','3f6fba83-a8e1-4105-bbf9-9b6c0c926a99','5.1.0','https://download.example.com/acemapper_v1.0.0.zip');
INSERT INTO product VALUES(2,'AcePrint','58e97b70-5703-41b9-b3fb-da5a8d8b4d22','1.5.0','https://download.example.com/aceprint_v1.0.0.zip');
-- TABLE registration
INSERT INTO registration VALUES(1,1,'2025-11-30','z9xiJQpG4FV5c6OUlptfW1/3GEY=','2023-01-01','2025-07-01','5.0.0');
-- TABLE license
INSERT INTO license VALUES(1,1,'ac7aa088-32bd-4313-a8bd-45c6927b58bc',1,1,12,'2025-12-01','2026-11-30','2026-11-30','');
-- TABLE feature
INSERT INTO feature VALUES(1,1,'PartTypes',0,'','999999999');
INSERT INTO feature VALUES(2,1,'Legacy',2,'True|False','False');
INSERT INTO feature VALUES(3,1,'Structured',2,'True|False','True');
INSERT INTO feature VALUES(4,2,'CatalogCount',0,'','1');
-- TABLE license_feature
COMMIT;
