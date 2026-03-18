-- ClawHost Database Initialization Script

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database user if not exists (for production)
-- DO
-- $do$
-- BEGIN
--    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'clawhost_user') THEN
--       CREATE ROLE clawhost_user LOGIN PASSWORD 'your_production_password';
--    END IF;
-- END
-- $do$;

-- Grant necessary permissions
-- GRANT CONNECT ON DATABASE clawhost_db TO clawhost_user;
-- GRANT USAGE ON SCHEMA public TO clawhost_user;
-- GRANT CREATE ON SCHEMA public TO clawhost_user;

-- Create indexes for performance (will be created by GORM migrations)
-- These are just placeholders for manual optimization if needed