-- SyncGuard MFA — Initial Schema
-- Migration: 001_initial_schema
-- Direction: UP

BEGIN;

-- ─── Extensions ──────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users ───────────────────────────────────────────────────────────────────
CREATE TYPE user_status AS ENUM ('active', 'disabled', 'locked');
CREATE TYPE user_source AS ENUM ('local', 'ldap', 'adfs');

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      TEXT NOT NULL,
    email         TEXT NOT NULL,
    display_name  TEXT NOT NULL DEFAULT '',
    status        user_status NOT NULL DEFAULT 'active',
    source        user_source NOT NULL DEFAULT 'local',
    password_hash TEXT NOT NULL DEFAULT '',   -- bcrypt; empty for ldap users
    failed_logins INTEGER NOT NULL DEFAULT 0,
    locked_until  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique    UNIQUE (email)
);

CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_status   ON users (status);

-- ─── TOTP Secrets ────────────────────────────────────────────────────────────
-- Secret stored AES-256-GCM encrypted; key managed via SYNCGUARD_ENCRYPTION_KEY
CREATE TABLE totp_secrets (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret_enc   BYTEA NOT NULL,             -- encrypted base32 secret
    enrolled     BOOLEAN NOT NULL DEFAULT FALSE,
    enrolled_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT totp_secrets_user_unique UNIQUE (user_id)
);

-- ─── Devices ─────────────────────────────────────────────────────────────────
CREATE TYPE device_status AS ENUM (
    'unknown',      -- first seen, no decision yet
    'pending',      -- MFA triggered, awaiting approval
    'approved',     -- trusted, allowed through
    'blocked',      -- explicitly denied
    'quarantine'    -- isolated, under review
);

CREATE TABLE devices (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_eas_id  TEXT NOT NULL,            -- DeviceId from EAS header
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_type    TEXT NOT NULL DEFAULT '', -- iPhone, Android, etc.
    device_model   TEXT NOT NULL DEFAULT '',
    user_agent     TEXT NOT NULL DEFAULT '',
    last_ip        INET,
    status         device_status NOT NULL DEFAULT 'unknown',
    trust_token    TEXT,                     -- opaque token stored on gateway
    trusted_at     TIMESTAMPTZ,
    trust_expiry   TIMESTAMPTZ,
    enrolled_at    TIMESTAMPTZ,
    blocked_at     TIMESTAMPTZ,
    block_reason   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT devices_eas_user_unique UNIQUE (device_eas_id, user_id)
);

CREATE INDEX idx_devices_user_id     ON devices (user_id);
CREATE INDEX idx_devices_eas_id      ON devices (device_eas_id);
CREATE INDEX idx_devices_status      ON devices (status);
CREATE INDEX idx_devices_trust_token ON devices (trust_token) WHERE trust_token IS NOT NULL;

-- ─── MFA Challenges ──────────────────────────────────────────────────────────
-- Authoritative record of challenge lifecycle (fast path is Redis)
CREATE TYPE challenge_method AS ENUM ('totp', 'email', 'push', 'sms');
CREATE TYPE challenge_status AS ENUM ('pending', 'completed', 'expired', 'failed', 'locked');

CREATE TABLE mfa_challenges (
    id            TEXT PRIMARY KEY,          -- UUID string, also Redis key
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_eas_id TEXT NOT NULL,
    method        challenge_method NOT NULL DEFAULT 'totp',
    status        challenge_status NOT NULL DEFAULT 'pending',
    attempts      INTEGER NOT NULL DEFAULT 0,
    otp_hash      TEXT,                      -- bcrypt hash of email/SMS OTP
    client_ip     INET,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    completed_at  TIMESTAMPTZ
);

CREATE INDEX idx_challenges_user_id    ON mfa_challenges (user_id);
CREATE INDEX idx_challenges_created_at ON mfa_challenges (created_at);
CREATE INDEX idx_challenges_status     ON mfa_challenges (status);

-- ─── Audit Events ────────────────────────────────────────────────────────────
CREATE TYPE audit_severity AS ENUM ('info', 'warn', 'error', 'critical');

CREATE TABLE audit_events (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action        TEXT NOT NULL,             -- AUTH_ATTEMPT, MFA_SUCCESS, etc.
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    username      TEXT NOT NULL DEFAULT '',
    device_eas_id TEXT NOT NULL DEFAULT '',
    client_ip     INET,
    challenge_id  TEXT NOT NULL DEFAULT '',
    result        TEXT NOT NULL DEFAULT '',  -- SUCCESS, FAILURE, DENIED, etc.
    details       JSONB,
    severity      audit_severity NOT NULL DEFAULT 'info'
);

CREATE INDEX idx_audit_occurred_at   ON audit_events (occurred_at DESC);
CREATE INDEX idx_audit_user_id       ON audit_events (user_id);
CREATE INDEX idx_audit_action        ON audit_events (action);
CREATE INDEX idx_audit_severity      ON audit_events (severity);
CREATE INDEX idx_audit_details       ON audit_events USING GIN (details);

-- ─── Admin Sessions ──────────────────────────────────────────────────────────
CREATE TABLE admin_sessions (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    ip         INET,
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    CONSTRAINT admin_sessions_token_unique UNIQUE (token_hash)
);

CREATE INDEX idx_admin_sessions_user_id    ON admin_sessions (user_id);
CREATE INDEX idx_admin_sessions_expires_at ON admin_sessions (expires_at);

-- ─── System Config ────────────────────────────────────────────────────────────
-- Key/value store for runtime-editable policy configuration
CREATE TABLE system_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

INSERT INTO system_config (key, value) VALUES
    ('mfa.default_method',        'totp'),
    ('mfa.challenge_ttl_seconds', '300'),
    ('mfa.max_attempts',          '5'),
    ('device.trust_ttl_hours',    '720'),
    ('device.auto_enrol',         'false'),
    ('gateway.fail_closed',       'true');

-- ─── Updated-at trigger ──────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at        BEFORE UPDATE ON users        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_totp_secrets_updated_at BEFORE UPDATE ON totp_secrets FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_devices_updated_at      BEFORE UPDATE ON devices      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
