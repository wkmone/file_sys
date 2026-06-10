package main

const migrationUsers = `
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    display_name    VARCHAR(128) NOT NULL,
    avatar_url      VARCHAR(512),
    role            VARCHAR(32)  NOT NULL DEFAULT 'member'
                        CHECK (role IN ('super_admin', 'admin', 'member')),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`

const migrationTeams = `
CREATE TABLE IF NOT EXISTS teams (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS team_members (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id         UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(32) NOT NULL DEFAULT 'member'
                        CHECK (role IN ('owner', 'admin', 'member')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team ON team_members(team_id);
`

const migrationFolders = `
CREATE TABLE IF NOT EXISTS folders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    parent_id       UUID REFERENCES folders(id) ON DELETE CASCADE,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id         UUID REFERENCES teams(id) ON DELETE SET NULL,
    folder_path     LTREE NOT NULL DEFAULT '',
    is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name, parent_id, owner_id)
);

CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_folders_owner ON folders(owner_id);
CREATE INDEX IF NOT EXISTS idx_folders_team ON folders(team_id);
CREATE INDEX IF NOT EXISTS idx_folders_path ON folders USING GIST (folder_path);
`

const migrationFiles = `
CREATE TABLE IF NOT EXISTS files (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    original_name   VARCHAR(255) NOT NULL,
    folder_id       UUID REFERENCES folders(id) ON DELETE SET NULL,
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id         UUID REFERENCES teams(id) ON DELETE SET NULL,
    mime_type       VARCHAR(128) NOT NULL,
    file_size       BIGINT NOT NULL DEFAULT 0,
    file_ext        VARCHAR(16) NOT NULL DEFAULT '',
    storage_key     VARCHAR(512) NOT NULL,
    content_hash    VARCHAR(64),
    current_version INT NOT NULL DEFAULT 1,
    is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name, folder_id, owner_id)
);

CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id);
CREATE INDEX IF NOT EXISTS idx_files_owner ON files(owner_id);
CREATE INDEX IF NOT EXISTS idx_files_team ON files(team_id);
CREATE INDEX IF NOT EXISTS idx_files_ext ON files(file_ext);
CREATE INDEX IF NOT EXISTS idx_files_name_trgm ON files USING GIN (name gin_trgm_ops);
`

const migrationFileVersions = `
CREATE TABLE IF NOT EXISTS file_versions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id         UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    version_number  INT NOT NULL,
    storage_key     VARCHAR(512) NOT NULL,
    file_size       BIGINT NOT NULL DEFAULT 0,
    content_hash    VARCHAR(64),
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    change_note     VARCHAR(512),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (file_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_versions_file ON file_versions(file_id);
`

const migrationPermissions = `
CREATE TABLE IF NOT EXISTS permissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    folder_id       UUID REFERENCES folders(id) ON DELETE CASCADE,
    file_id         UUID REFERENCES files(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    team_id         UUID REFERENCES teams(id) ON DELETE CASCADE,
    permission      VARCHAR(32) NOT NULL DEFAULT 'read'
                        CHECK (permission IN ('read', 'write', 'admin')),
    granted_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_grantee CHECK (
        (user_id IS NOT NULL AND team_id IS NULL) OR
        (user_id IS NULL AND team_id IS NOT NULL)
    ),
    CONSTRAINT chk_resource CHECK (
        (folder_id IS NOT NULL AND file_id IS NULL) OR
        (folder_id IS NULL AND file_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_unique_user_folder
    ON permissions(user_id, folder_id) WHERE user_id IS NOT NULL AND folder_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_unique_team_folder
    ON permissions(team_id, folder_id) WHERE team_id IS NOT NULL AND folder_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_unique_user_file
    ON permissions(user_id, file_id) WHERE user_id IS NOT NULL AND file_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_unique_team_file
    ON permissions(team_id, file_id) WHERE team_id IS NOT NULL AND file_id IS NOT NULL;
`

const migrationRefreshTokens = `
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    device_info     VARCHAR(255),
    ip_address      INET,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
`

const migrationOOSessions = `
CREATE TABLE IF NOT EXISTS onlyoffice_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id         UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_key    VARCHAR(128) NOT NULL,
    callback_url    VARCHAR(512),
    mode            VARCHAR(16) NOT NULL DEFAULT 'edit'
                        CHECK (mode IN ('edit', 'view', 'comment', 'review', 'fillForms')),
    status          VARCHAR(32) NOT NULL DEFAULT 'opened',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oo_sessions_file ON onlyoffice_sessions(file_id);
CREATE INDEX IF NOT EXISTS idx_oo_sessions_key ON onlyoffice_sessions(document_key);
`

const migrationOOSessionsModeUpdate = `
DO $$ BEGIN
	ALTER TABLE onlyoffice_sessions DROP CONSTRAINT IF EXISTS onlyoffice_sessions_mode_check;
EXCEPTION WHEN undefined_table THEN
	-- table doesn't exist yet, skip
END $$;
ALTER TABLE onlyoffice_sessions ADD CONSTRAINT onlyoffice_sessions_mode_check CHECK (mode IN ('edit', 'view', 'comment', 'review', 'fillForms'));
`

const migrationJoinRequests = `
CREATE TABLE IF NOT EXISTS join_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id         UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(16) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_join_requests_team ON join_requests(team_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_user ON join_requests(user_id);
`
