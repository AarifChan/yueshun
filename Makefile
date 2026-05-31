.PHONY: build dev swagger migrate lint test clean help

# 变量
APP_NAME := zhizhang-server
MAIN_FILE := cmd/server/main.go
BUILD_DIR := build
CONFIG_FILE := configs/config.yaml

# 默认目标
help: ## 显示帮助信息
	@echo "可用的命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

dev: ## 开发模式运行（热重载需安装 air）
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed, running with go run..."; \
		go run $(MAIN_FILE); \
	fi

run: ## 直接运行（不调 air）
	go run $(MAIN_FILE)

swagger: ## 生成 Swagger 文档
	@echo "Generating swagger docs..."
	@which swag >/dev/null 2>&1 || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	$(HOME)/go/bin/swag init -g $(MAIN_FILE) -o docs
	@echo "Swagger docs generated at docs/"

migrate: ## 数据库迁移（自动迁移模式）
	@echo "Running database migration..."
	go run $(MAIN_FILE) --migrate

lint: ## 代码检查
	@echo "Running linter..."
	@which golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

test: ## 运行测试
	@echo "Running tests..."
	go test -v -race -cover ./...

test-short: ## 运行短测试
	go test -v -short ./...

deps: ## 安装依赖
	go mod download
	go mod tidy

clean: ## 清理构建产物
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 数据库相关
pg-up: ## 启动本地 PostgreSQL (Docker)
	@docker run -d --name zhizhang-pg \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=zhizhang \
		-p 5432:5432 \
		-v zhizhang-pg-data:/var/lib/postgresql/data \
		postgres:15-alpine || echo "Container may already exist, starting..."
	@docker start zhizhang-pg 2>/dev/null || true
	@echo "PostgreSQL ready at localhost:5432"

pg-down: ## 停止 PostgreSQL
	@docker stop zhizhang-pg 2>/dev/null || true

pg-logs: ## 查看 PostgreSQL 日志
	@docker logs -f zhizhang-pg

redis-up: ## 启动本地 Redis (Docker)
	@docker run -d --name zhizhang-redis \
		-p 6379:6379 \
		redis:7-alpine || echo "Container may already exist, starting..."
	@docker start zhizhang-redis 2>/dev/null || true
	@echo "Redis ready at localhost:6379"

redis-down: ## 停止 Redis
	@docker stop zhizhang-redis 2>/dev/null || true

infra-up: ## 启动所有基础设施
	@$(MAKE) pg-up
	@$(MAKE) redis-up
	@echo "All infrastructure ready"

infra-down: ## 停止所有基础设施
	@$(MAKE) pg-down
	@$(MAKE) redis-down
