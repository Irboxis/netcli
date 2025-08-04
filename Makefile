# Makefile for Cross-Compiling net-cli

# 项目名称
PROJECT_NAME := nctl

# 主入口文件路径
MAIN_PACKAGE := ./cmd

# 输出目录
BUILD_DIR := build

# 获取 Go 版本
GO_VERSION := $(shell go version | awk '{print $$3}' | sed 's/go//')

.PHONY: all clean build_linux build_windows build_darwin build_all

# 默认目标：构建所有支持的平台
all: build_all

# 清理构建目录
clean:
	@echo "Cleaning up build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# 构建所有支持的平台
build_all:
	@mkdir -p $(BUILD_DIR)
	$(MAKE) linux
	$(MAKE) windows

# --- Linux 构建目标 ---
# 采用 CentOS 6/7/8/9/10 系列，选择 amd64 架构
# 需要向下兼容到 CentOS 6 ，可能会有 libc 等兼容性问题，使用 CGO_ENABLED=0 进行纯静态编译
# CentOS 7+ 兼容性更好，但纯静态编译是更保险的选择
linux:
	@echo "Building for Linux (RHEL/CentOS/Ubuntu/Debian)..."
	# Linux AMD64 (64-bit)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(PROJECT_NAME)_linux_amd64 $(MAIN_PACKAGE)
	# Linux 386 (32-bit)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(PROJECT_NAME)_linux_386 $(MAIN_PACKAGE)
	@echo "Linux builds complete."

# --- Windows 构建目标 ---
# Windows 7/10/11 选择 amd64 架构，但 32-bit (386) 也可能存在
windows:
	@echo "Building for Windows..."
	# Windows AMD64 (64-bit)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(PROJECT_NAME)_windows_amd64.exe $(MAIN_PACKAGE)
	# Windows 386 (32-bit)
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(PROJECT_NAME)_windows_386.exe $(MAIN_PACKAGE)
	@echo "Windows builds complete."