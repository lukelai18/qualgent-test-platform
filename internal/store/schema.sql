-- Database schema for QualGent Test Platform

-- Jobs table - stores individual test jobs
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id TEXT NOT NULL,
    app_version_id TEXT NOT NULL,
    test_path TEXT NOT NULL,
    priority INTEGER DEFAULT 0,
    target TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    job_group_id UUID,
    idempotency_key TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Job groups table - groups jobs by app_version_id and target
CREATE TABLE IF NOT EXISTS job_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_version_id TEXT NOT NULL,
    target TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'SCHEDULED',
    agent_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Agents table - stores execution agents
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname TEXT UNIQUE NOT NULL,
    target_capability TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'IDLE',
    last_heartbeat_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add foreign key constraints
ALTER TABLE jobs ADD CONSTRAINT fk_jobs_job_group_id FOREIGN KEY (job_group_id) REFERENCES job_groups(id);
ALTER TABLE job_groups ADD CONSTRAINT fk_job_groups_agent_id FOREIGN KEY (agent_id) REFERENCES agents(id);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_jobs_status_priority ON jobs(status, priority);
CREATE INDEX IF NOT EXISTS idx_jobs_app_version_id ON jobs(app_version_id);
CREATE INDEX IF NOT EXISTS idx_jobs_job_group_id ON jobs(job_group_id);
CREATE INDEX IF NOT EXISTS idx_job_groups_status ON job_groups(status);
CREATE INDEX IF NOT EXISTS idx_job_groups_app_version_target ON job_groups(app_version_id, target);
CREATE INDEX IF NOT EXISTS idx_agents_status_capability ON agents(status, target_capability);
CREATE INDEX IF NOT EXISTS idx_agents_last_heartbeat ON agents(last_heartbeat_at);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language plpgsql;

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_groups_updated_at BEFORE UPDATE ON job_groups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 