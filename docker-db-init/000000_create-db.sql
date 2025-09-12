SELECT 'CREATE DATABASE sentinel' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sentinel')\gexec
CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD '1234'; -- TODO change in prod
