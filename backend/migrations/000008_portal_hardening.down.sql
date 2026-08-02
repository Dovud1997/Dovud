DROP INDEX IF EXISTS idx_documents_customer;
ALTER TABLE documents DROP COLUMN IF EXISTS customer_id;
