#!/usr/bin/env bash
# 同时启动后端 (Go) 和前端 (Vite)
# 用法: scripts/start.sh [--build]   (--build: 先编译后端再运行，否则 go run)

set -e
cd "$(dirname "$0")/.."

USE_BUILD=false
[ "$1" = "--build" ] && USE_BUILD=true

# 递归结束进程树（go run / npm 都会派生子进程，只杀父进程会留孤儿）
kill_tree() {
    local pid=$1 child
    for child in $(pgrep -P "$pid" 2>/dev/null); do
        kill_tree "$child"
    done
    kill "$pid" 2>/dev/null || true
}

# 存活性判断：排除僵尸进程（子进程退出后未被 wait 前 kill -0 仍为真）
alive() {
    kill -0 "$1" 2>/dev/null || return 1
    ! ps -o stat= -p "$1" 2>/dev/null | grep -q Z
}

# 释放被旧进程占用的端口（上次未正常退出的残留）
free_port() {
    local port=$1 pids
    pids=$(lsof -ti :"$port" -sTCP:LISTEN 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "==> 端口 $port 被进程 $(echo $pids | tr '\n' ' ')占用，已停止旧进程"
        kill $pids 2>/dev/null || true
        sleep 1
        kill -9 $pids 2>/dev/null || true
    fi
}

if [ "$USE_BUILD" = true ]; then
    echo "==> 编译后端..."
    make build
    BACKEND_CMD="./build/zhizhang-server"
else
    BACKEND_CMD="go run cmd/server/main.go"
fi

free_port 8388
free_port 3000

# 启动后端：若启动失败（如端口在检查后被抢占），杀掉占用者并重试，最多 3 次
BACKEND_PID=""
for attempt in 1 2 3; do
    echo "==> 启动后端: $BACKEND_CMD"
    $BACKEND_CMD &
    BACKEND_PID=$!
    # 等待后端完成初始化（redis 重试会拖慢几秒），确认存活即成功
    sleep 6
    if alive "$BACKEND_PID"; then
        break
    fi
    wait "$BACKEND_PID" 2>/dev/null || true
    echo "==> 后端启动失败，释放端口后重试 ($attempt/3)..."
    free_port 8388
done
if ! alive "$BACKEND_PID"; then
    echo "==> 后端多次启动失败，退出"
    exit 1
fi

echo "==> 启动前端 (vite)..."
(cd frontend && npm run dev -- --strictPort) &
FRONTEND_PID=$!

echo ""
echo "后端 PID: $BACKEND_PID, 前端 PID: $FRONTEND_PID"
echo "按 Ctrl+C 停止前后端"

cleanup() {
    echo ""
    echo "==> 停止进程..."
    kill_tree "$BACKEND_PID"
    kill_tree "$FRONTEND_PID"
    exit 0
}
trap cleanup INT TERM

# macOS 自带 bash 3.2 不支持 wait -n，用轮询等待任一进程退出
EXIT_CODE=0
while alive "$BACKEND_PID" && alive "$FRONTEND_PID"; do
    sleep 1
done
if ! alive "$BACKEND_PID"; then
    wait "$BACKEND_PID" || EXIT_CODE=$?
else
    wait "$FRONTEND_PID" || EXIT_CODE=$?
fi
echo "==> 一个进程已退出 (code=$EXIT_CODE)，停止另一个..."
kill_tree "$BACKEND_PID"
kill_tree "$FRONTEND_PID"
exit "$EXIT_CODE"
