-- =============================================================================
-- PostgreSQL Production Setup Script
-- =============================================================================
-- PURPOSE : Create database, schema, and app runtime user
--           with least-privilege access controls.
-- USAGE   : Run via setup_db.sh (reads credentials from .envrc)
--           OR manually: psql -U anpks \
--                          -d postgres \
--                          -v db_name=myapp \
--                          -v app_user=app_user \
--                          -v app_password=secret \
--                          -f postgres_setup.sql
-- =============================================================================


-- =============================================================================
-- SECTION 1: CREATE DATABASE
-- =============================================================================

CREATE DATABASE :db_name
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE   = 'en_US.UTF-8'
    TEMPLATE   = template0;

-- Connect to the new database before running the rest
\c :db_name


-- =============================================================================
-- SECTION 2: CREATE USER
-- =============================================================================

CREATE USER :app_user WITH
    PASSWORD :'app_password'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    LOGIN;


-- =============================================================================
-- SECTION 3: DATABASE-LEVEL PRIVILEGES
-- =============================================================================

-- Revoke default public access
REVOKE ALL ON DATABASE :db_name FROM PUBLIC;

-- User need to connect
GRANT CONNECT ON DATABASE :db_name TO :app_user;
GRANT CREATE ON DATABASE :db_name TO :app_user;

-- =============================================================================
-- SECTION 4: PUBLIC SCHEMA PRIVILEGES
-- =============================================================================

-- lock down public schema from everyone
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

-- grant only to app user
GRANT USAGE  ON SCHEMA public TO :app_user;
GRANT CREATE ON SCHEMA public TO :app_user;

-- =============================================================================
-- SECTION 5: TABLE-LEVEL PRIVILEGES (existing tables)
-- =============================================================================

GRANT ALL ON ALL TABLES IN SCHEMA public TO :app_user;

-- Sequences (required for SERIAL / BIGSERIAL / IDENTITY columns)
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO :app_user;


-- =============================================================================
-- SECTION 6: DEFAULT PRIVILEGES (for future tables created by migrations)
-- =============================================================================

ALTER DEFAULT PRIVILEGES FOR USER :app_user IN SCHEMA public
    GRANT ALL ON TABLES TO :app_user;

ALTER DEFAULT PRIVILEGES FOR USER :app_user IN SCHEMA public
    GRANT ALL ON SEQUENCES TO :app_user;


-- =============================================================================
-- SECTION 7: VERIFY
-- =============================================================================

SELECT rolname, rolsuper, rolcreatedb, rolcreaterole, rolcanlogin
FROM pg_roles
WHERE rolname IN (:'app_user');

\dn+

-- =============================================================================
-- END OF SCRIPT
-- =============================================================================
