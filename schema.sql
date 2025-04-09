-- Database schema for HSOETLNLM (PostgreSQL)

-- Connections Table: Stores details for source and target systems
CREATE TABLE Connections (
    ID BIGSERIAL PRIMARY KEY, -- Use BIGSERIAL for auto-incrementing BIGINT
    Name VARCHAR(255) NOT NULL UNIQUE,
    Type VARCHAR(50) NOT NULL, -- e.g., 'oracle', 'sqlserver', 's3', 'snowflake'
    ConnectionString TEXT NOT NULL, -- Can store complex connection details
    CreatedAt TIMESTAMP NOT NULL DEFAULT NOW(), -- Use TIMESTAMP and NOW()
    UpdatedAt TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ReplicationTasks Table: Defines a data replication job
CREATE TABLE ReplicationTasks (
    ID BIGSERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL UNIQUE,
    SourceConnectionID BIGINT NOT NULL,
    TargetConnectionID BIGINT NOT NULL,
    Schedule VARCHAR(100) NULL, -- e.g., cron expression
    DataSelectionCriteria TEXT NULL, -- e.g., SQL query, S3 prefix
    TransformationRules TEXT NULL, -- e.g., Bloblang script
    TemporalWorkflowID VARCHAR(255) NULL,
    Status VARCHAR(50) NOT NULL, -- e.g., 'active', 'inactive', 'paused'
    CreatedAt TIMESTAMP NOT NULL DEFAULT NOW(),
    UpdatedAt TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Foreign Key constraints
    FOREIGN KEY (SourceConnectionID) REFERENCES Connections(ID),
    FOREIGN KEY (TargetConnectionID) REFERENCES Connections(ID)
);

-- ReplicationRuns Table: Stores history and status of task executions
CREATE TABLE ReplicationRuns (
    ID BIGSERIAL PRIMARY KEY,
    ReplicationTaskID BIGINT NOT NULL,
    StartTime TIMESTAMP NOT NULL,
    EndTime TIMESTAMP NULL, -- Nullable until the run completes
    Status VARCHAR(50) NOT NULL, -- e.g., 'loading', 'running', 'completed', 'failed'
    ErrorDetails TEXT NULL, -- Store error messages if the run failed
    TemporalRunID VARCHAR(255) NULL,
    CreatedAt TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Foreign Key constraint
    FOREIGN KEY (ReplicationTaskID) REFERENCES ReplicationTasks(ID) ON DELETE CASCADE -- Cascade delete if task is deleted
);

-- BenthosConfigurations Table: Stores reusable Benthos pipeline snippets or full configs
CREATE TABLE BenthosConfigurations (
    ID BIGSERIAL PRIMARY KEY,
    Name VARCHAR(255) NOT NULL UNIQUE,
    Configuration TEXT NOT NULL, -- Stores the Benthos YAML/JSON configuration
    CreatedAt TIMESTAMP NOT NULL DEFAULT NOW(),
    UpdatedAt TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Optional: Add Indexes for common lookups
CREATE INDEX IX_ReplicationTasks_SourceConnectionID ON ReplicationTasks(SourceConnectionID);
CREATE INDEX IX_ReplicationTasks_TargetConnectionID ON ReplicationTasks(TargetConnectionID);
CREATE INDEX IX_ReplicationRuns_ReplicationTaskID ON ReplicationRuns(ReplicationTaskID);
CREATE INDEX IX_ReplicationRuns_Status ON ReplicationRuns(Status);

-- Note: Syntax for IDENTITY, DEFAULT GETDATE(), TIMESTAMP might vary slightly depending on the specific SQL database (e.g., PostgreSQL, MySQL). Adjust as needed. 