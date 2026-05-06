-- SyncGuard MFA — Initial Schema
-- Migration: 001_initial_schema
-- Direction: DOWN

BEGIN;

DROP TABLE IF EXISTS system_config     CASCADE;
DROP TABLE IF EXISTS admin_sessions    CASCADE;
DROP TABLE IF EXISTS audit_events      CASCADE;
DROP TABLE IF EXISTS mfa_challenges    CASCADE;
DROP TABLE IF EXISTS devices           CASCADE;
DROP TABLE IF EXISTS totp_secrets      CASCADE;
DROP TABLE IF EXISTS users             CASCADE;

DROP FUNCTION IF EXISTS set_updated_at CASCADE;

DROP TYPE IF EXISTS audit_severity   CASCADE;
DROP TYPE IF EXISTS challenge_status CASCADE;
DROP TYPE IF EXISTS challenge_method CASCADE;
DROP TYPE IF EXISTS device_status    CASCADE;
DROP TYPE IF EXISTS user_source      CASCADE;
DROP TYPE IF EXISTS user_status      CASCADE;

DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";

COMMIT;
