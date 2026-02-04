-- GM Server 数据库初始化SQL脚本
-- 可以直接使用 psql 执行此脚本

-- 创建数据库（如果不存在）
-- 注意：需要在 postgres 数据库中执行
SELECT 'CREATE DATABASE gm_server'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'gm_server')\gexec

-- 连接到新创建的数据库
\c gm_server

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 显示成功信息
SELECT 'Database gm_server initialized successfully!' AS message;
