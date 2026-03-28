CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(32) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT true,
    company_id VARCHAR(64),
    branch_id VARCHAR(64),
    onboarding_completed BOOLEAN DEFAULT false,
    icp_certificate_chain TEXT,
    icp_certificate_serial VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_branch_id ON users(branch_id);

CREATE TABLE IF NOT EXISTS auditor_access_grants (
    id VARCHAR(64) PRIMARY KEY,
    auditor_id VARCHAR(64) NOT NULL,
    scope VARCHAR(32) NOT NULL,
    inventory_id VARCHAR(64),
    company_branch_id VARCHAR(64),
    company_id VARCHAR(64),
    granted_by VARCHAR(64) NOT NULL,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auditor_access_auditor ON auditor_access_grants(auditor_id);
CREATE INDEX idx_auditor_access_company ON auditor_access_grants(company_id);

CREATE TABLE IF NOT EXISTS known_companies (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_tokens (
    user_id VARCHAR(64) NOT NULL,
    token VARCHAR(512) NOT NULL,
    platform VARCHAR(16) NOT NULL,
    registered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, token)
);

CREATE INDEX idx_device_tokens_user ON device_tokens(user_id);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_outbox_unsent ON outbox(sent, created_at);
