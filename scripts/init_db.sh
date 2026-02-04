#!/bin/bash

# 数据库一键初始化脚本
# 用于创建数据库和用户

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-gm_server}"
DB_PASSWORD="${DB_PASSWORD:-}"

echo -e "${GREEN}=== GM Server 数据库初始化脚本 ===${NC}"
echo ""

# 检查psql是否安装
if ! command -v psql &> /dev/null; then
    echo -e "${RED}错误: 未找到 psql 命令，请先安装 PostgreSQL 客户端${NC}"
    exit 1
fi

# 提示输入数据库管理员密码
if [ -z "$PGPASSWORD" ] && [ -z "$DB_PASSWORD" ]; then
    echo -e "${YELLOW}请输入 PostgreSQL 管理员密码（用于创建数据库和用户）:${NC}"
    read -s PGPASSWORD
    export PGPASSWORD
fi

# 设置连接参数
export PGHOST=$DB_HOST
export PGPORT=$DB_PORT
export PGUSER=$DB_USER

echo -e "${GREEN}连接信息:${NC}"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  用户: $DB_USER"
echo "  数据库: $DB_NAME"
echo ""

# 检查数据库是否已存在
if psql -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo -e "${YELLOW}数据库 '$DB_NAME' 已存在${NC}"
    read -p "是否删除并重新创建? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}删除数据库 '$DB_NAME'...${NC}"
        psql -c "DROP DATABASE IF EXISTS $DB_NAME;"
        echo -e "${GREEN}数据库已删除${NC}"
    else
        echo -e "${GREEN}跳过数据库创建${NC}"
        exit 0
    fi
fi

# 创建数据库
echo -e "${GREEN}创建数据库 '$DB_NAME'...${NC}"
psql -c "CREATE DATABASE $DB_NAME;"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}数据库创建成功${NC}"
else
    echo -e "${RED}数据库创建失败${NC}"
    exit 1
fi

# 创建扩展（如果需要）
echo -e "${GREEN}创建数据库扩展...${NC}"
psql -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" || true

echo ""
echo -e "${GREEN}=== 数据库初始化完成 ===${NC}"
echo -e "${GREEN}数据库名称: $DB_NAME${NC}"
echo -e "${GREEN}现在可以启动服务，服务会自动执行表结构迁移${NC}"
echo ""
