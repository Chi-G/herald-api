-- Herald local development seed script

-- 1. Create default test tenant (e.g. AuraMed)
INSERT INTO tenants (id, name, slug)
VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'AuraMed Dev', 'auramed-dev')
ON CONFLICT (slug) DO NOTHING;

-- 2. Create test API Key
-- Raw API Key: hrld_live_testkey123
-- SHA-256 Hash of "hrld_live_testkey123": 9b769213192080313c01c0f4f9f7435f3dfd96ebf4b005a76c8c97ec59b66f24
INSERT INTO api_keys (tenant_id, name, key_hash, key_prefix)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'AuraMed Development Key',
    'b9881ad0c194b5ee9d40554721ce15eb3d8d4eb81ac439729286ec4830b0ae9f',
    'hrld_live_t'
)
ON CONFLICT (key_hash) DO NOTHING;

