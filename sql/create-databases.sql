SELECT 'CREATE DATABASE test_db_1'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test_db_1')\gexec

SELECT 'CREATE DATABASE test_db_2'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test_db_2')\gexec

SELECT 'CREATE DATABASE test_db_3'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test_db_3')\gexec

SELECT 'CREATE DATABASE test_db_4'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test_db_4')\gexec