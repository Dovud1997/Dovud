ALTER TABLE documents ADD COLUMN IF NOT EXISTS customer_id UUID NULL;
CREATE INDEX IF NOT EXISTS idx_documents_customer ON documents(tenant_id, customer_id);
