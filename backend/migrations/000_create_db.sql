-- ==========================================
-- 建库脚本 (以 postgres 超级用户执行)
-- psql -U postgres -f 000_create_db.sql
-- ==========================================

CREATE DATABASE file_sys
    WITH ENCODING 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE template0;

-- 如果库已存在会报错，忽略即可:
-- ERROR: database "file_sys" already exists
