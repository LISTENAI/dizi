#!/bin/bash

# Dizi MCP Server Installation Script

set -e

APP_NAME="dizi"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/dizi"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go first."
        exit 1
    fi
    
    log_info "Go version: $(go version)"
}

# 构建应用
build_app() {
    log_info "Building $APP_NAME..."
    
    if [ -f "Makefile" ]; then
        make build
    else
        go build -ldflags="-s -w" -o $APP_NAME
    fi
    
    if [ ! -f "$APP_NAME" ]; then
        log_error "Build failed"
        exit 1
    fi
    
    log_success "Build completed"
}

# 安装二进制文件
install_binary() {
    log_info "Installing $APP_NAME to $INSTALL_DIR..."
    
    # 检查是否需要 sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        log_warn "Need sudo permissions to install to $INSTALL_DIR"
        sudo cp "$APP_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$APP_NAME"
    else
        cp "$APP_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$APP_NAME"
    fi
    
    log_success "$APP_NAME installed to $INSTALL_DIR"
}

# 设置配置目录
setup_config() {
    log_info "Setting up configuration..."
    
    # 创建配置目录
    mkdir -p "$CONFIG_DIR"
    
    # 复制示例配置文件
    if [ -f "dizi.example.yml" ] && [ ! -f "$CONFIG_DIR/dizi.yml" ]; then
        cp "dizi.example.yml" "$CONFIG_DIR/dizi.yml"
        log_success "Example configuration copied to $CONFIG_DIR/dizi.yml"
        log_info "Please edit $CONFIG_DIR/dizi.yml to customize your configuration"
    fi
    
    # 创建本地配置文件链接
    if [ -f "dizi.yml" ]; then
        log_info "Local configuration file found: dizi.yml"
    elif [ ! -f "dizi.yml" ] && [ -f "$CONFIG_DIR/dizi.yml" ]; then
        ln -sf "$CONFIG_DIR/dizi.yml" "dizi.yml"
        log_info "Created symlink to global configuration"
    fi
}

# 验证安装
verify_installation() {
    log_info "Verifying installation..."
    
    if command -v $APP_NAME &> /dev/null; then
        log_success "$APP_NAME is installed and available in PATH"
        log_info "Version: $($APP_NAME -help | head -n 1)"
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

# 显示使用说明
show_usage() {
    log_success "Installation completed successfully!"
    echo
    log_info "Quick start:"
    echo "  1. Edit configuration: ${CONFIG_DIR}/dizi.yml"
    echo "  2. Run server: $APP_NAME"
    echo "  3. For help: $APP_NAME -help"
    echo
    log_info "Configuration locations:"
    echo "  - Global: ${CONFIG_DIR}/dizi.yml"
    echo "  - Local: ./dizi.yml (current directory)"
    echo
    log_info "Examples:"
    echo "  $APP_NAME                    # Start with SSE transport (default)"
    echo "  $APP_NAME -port=9000         # Use custom port"
    echo "  $APP_NAME -transport=stdio   # Use stdio transport"
}

# 主函数
main() {
    log_info "Starting $APP_NAME installation..."
    
    check_dependencies
    build_app
    install_binary
    setup_config
    verify_installation
    show_usage
}

# 错误处理
trap 'log_error "Installation failed"; exit 1' ERR

# 检查脚本参数
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --uninstall    Uninstall $APP_NAME"
    exit 0
fi

if [ "$1" = "--uninstall" ]; then
    log_info "Uninstalling $APP_NAME..."
    
    if [ -f "$INSTALL_DIR/$APP_NAME" ]; then
        if [ ! -w "$INSTALL_DIR" ]; then
            sudo rm -f "$INSTALL_DIR/$APP_NAME"
        else
            rm -f "$INSTALL_DIR/$APP_NAME"
        fi
        log_success "$APP_NAME uninstalled from $INSTALL_DIR"
    else
        log_warn "$APP_NAME not found in $INSTALL_DIR"
    fi
    
    log_info "Configuration files in $CONFIG_DIR are preserved"
    log_info "Remove manually if needed: rm -rf $CONFIG_DIR"
    exit 0
fi

# 运行主函数
main