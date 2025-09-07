SELECT 'CREATE DATABASE sentinel' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sentinel')\gexec

