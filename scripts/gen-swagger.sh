#!/bin/bash
set -e

echo "Generating swagger docs..."

# 检查 swag 是否安装
if ! command -v swag &> /dev/null; then
    echo "swag not found, installing..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# 生成文档
swag init \
    --generalInfo cmd/server/main.go \
    --output docs \
    --parseDependency \
    --parseInternal

echo "Swagger docs generated at docs/"
echo "Visit: http://localhost:8388/swagger/index.html"
