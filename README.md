# Net CLI

这是一个用 Go 语言开发的命令行工具集，旨在提供一系列网络相关的实用命令，帮助用户方便地查询和管理网络接口及进行网络诊断。

## 📝 介绍 (Introduction)

`net-cli` 项目的目标是创建一个高效的、跨平台的网络命令行工具。

## ✨ 功能 (Features)

- **网络接口列表 (`nctl iface list`)**:
    - xxxx
    

## 🚀 快速开始 (Quick Start)

### 📦 安装 (Installation)

1.  **克隆仓库：**
    首先，你需要将本项目的代码克隆到你的本地机器上：
    ```bash
    git clone https://github.com/irboxis/net-cli.git
    cd net-cli
    ```

2.  **构建可执行文件：**
    进入项目根目录后，使用 make 命令构建可执行文件。构建后的文件将根据你的操作系统命名为 `nctl` 或 `nctl.exe`。
    ```bash
    make [all]
    ```
    *构建某一单独的平台：*
    ```bash
    make [net-windows|net-linux]
    ```

3.  **（可选）添加到 PATH 环境变量：**
    为了方便在任何目录下都能运行 `nctl` 命令，你可以将 `nctl` 可执行文件所在的目录添加到系统的 PATH 环境变量中。具体操作方法因操作系统而异。

### 🏃‍ 运行 (Usage)

1.  **自行构建可执行文件：**
    构建完成后，你可以在命令行中运行 `nctl` 命令。
2.  **选择发布版本直接运行：**
    开发中，尚未发布....

#### 1. 列出所有网络接口

要显示系统上所有网络接口的详细信息，只需运行：( 更多帮助信息，查看每条命令的 `-h` 或 `--help` 选项 )
```bash
./nctl iface list