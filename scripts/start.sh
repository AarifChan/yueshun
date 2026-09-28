#!/usr/bin/env bash
# 同时启动后端 (Go) 和前端 (Vite)
# 用法: scripts/start.sh [--build]   (--build: 先编译后端再运行，否则 go run)

set -e
cd "$(dirname "$0")/.."

USE_BUILD=false
[ "$1" = "--build" ] && USE_BUILD=true

if [ "$USE_BUILD" = true ]; then
    echo "==> 编译后端..."
    make build
    BACKEND_CMD="./build/zhizhang-server"
else
    BACKEND_CMD="go run cmd/server/main.go"
fi

echo "==> 启动后端: $BACKEND_CMD"
$BACKEND_CMD &
BACKEND_PID=$!

echo "==> 启动前端 (vite)..."
(cd frontend && npm run dev) &
FRONTEND_PID=$!

echo ""
echo "后端 PID: $BACKEND_PID, 前端 PID: $FRONTEND_PID"
echo "按 Ctrl+C 停止前后端"

cleanup() {
    echo ""
    echo "==> 停止进程..."
    kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
    exit 0
}
trap cleanup INT TERM

wait -n "$BACKEND_PID" "$FRONTEND_PID"
EXIT_CODE=$?
echo "==> 一个进程已退出 (code=$EXIT_CODE)，停止另一个..."
kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
exit "$EXIT_CODE"
