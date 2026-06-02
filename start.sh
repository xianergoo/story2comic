#!/bin/bash
set -e

cd "$(dirname "$0")"

# 首次运行自动创建 .env
if [ ! -f .env ]; then
  cp .env.example .env
  echo "已创建 .env，请编辑填入 SESSION_SECRET"
fi

# 编译
echo "Building..."
go build -o novelforge .

echo "Starting NovelForge on http://localhost:8080"
./novelforge
