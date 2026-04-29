# local-webdav

[English](#english) | [中文](#中文)

---

<a name="english"></a>
## English

A lightweight WebDAV server that maps virtual paths to local directories, providing a secure and convenient way to access your local files without exposing absolute paths.

### Features

- **Virtual Path Mapping**: Map local directories to clean virtual URLs (e.g., `http://localhost:43621/documents/`).
- **Multiple Shares**: Configure multiple local folders as separate WebDAV shares.
- **Authentication**: Optional username and password protection for each share.
- **Path Privacy**: Does not expose absolute file system paths to WebDAV clients.
- **Easy Configuration**: Simple TOML-based configuration.
- **Cross-Platform**: Built with Go, runs on macOS, Linux, and Windows.

### Installation

#### Using Go
```bash
go install github.com/hjgyuhuk/local-webdav@latest
```

#### From Source
```bash
git clone https://github.com/hjgyuhuk/local-webdav.git
cd local-webdav
go build -o local-webdav
```

### Quick Start

1.  **Initialize Config**: Run the server for the first time to generate the default configuration file.
    ```bash
    ./local-webdav serve
    ```
    This will create a config file at `~/.config/local-webdav/config.toml`.

2.  **Edit Config**: Open the generated TOML file and add your folders under `[[shares]]`.
    ```toml
    [server]
    address = "127.0.0.1:43621"

    [[shares]]
    name = "docs"
    path = "~/Documents"
    username = "admin"
    password = "yourpassword"

    [[shares]]
    name = "media"
    path = "/Volumes/External/Media"
    # No username/password means no authentication required
    ```

3.  **Start Server**: Run the serve command again.
    ```bash
    ./local-webdav serve
    ```

4.  **Access**: Connect your WebDAV client (Finder, Windows Explorer, etc.) to:
    - `http://127.0.0.1:43621/docs/`
    - `http://127.0.0.1:43621/media/`

### Chrome Extension (Cookie Monitor)

The project includes a companion Chrome extension that can automatically sync cookies from specified websites to your WebDAV server.

#### Features
- **Auto-Sync**: Automatically detects cookie changes and uploads them to WebDAV.
- **Netscape Format**: Exports cookies in the standard Netscape format (compatible with `curl`, `wget`, etc.).
- **Domain Filtering**: Only sync cookies for domains you specify.
- **Secure**: Credentials are stored locally in the extension.

#### Build & Install
1.  Navigate to the `extension` directory:
    ```bash
    cd extension
    pnpm install
    ```
2.  Build the extension:
    ```bash
    pnpm build
    ```
3.  Load the extension in Chrome:
    - Open `chrome://extensions/`
    - Enable "Developer mode"
    - Click "Load unpacked" and select the `extension/.output/chrome-mv3` folder.

---

<a name="中文"></a>
## 中文

一个轻量级的 WebDAV 服务器，可将虚拟路径映射到本地目录。提供了一种安全且便捷的方式来访问本地文件，同时不会泄露系统的绝对路径。

### 功能特性

- **虚拟路径映射**: 将本地目录映射为简洁的虚拟 URL（例如：`http://localhost:43621/documents/`）。
- **多共享支持**: 可配置多个本地文件夹作为独立的 WebDAV 共享。
- **身份认证**: 为每个共享提供可选的用户名和密码保护。
- **路径隐私**: 不向 WebDAV 客户端暴露文件系统的绝对路径。
- **简易配置**: 基于 TOML 的简单配置方式。
- **跨平台**: 使用 Go 编写，支持 macOS、Linux 和 Windows。

### 安装

#### 使用 Go
```bash
go install github.com/hjgyuhuk/local-webdav@latest
```

#### 源码编译
```bash
git clone https://github.com/hjgyuhuk/local-webdav.git
cd local-webdav
go build -o local-webdav
```

### 快速上手

1.  **初始化配置**: 首次运行服务器以生成默认配置文件。
    ```bash
    ./local-webdav serve
    ```
    这将在 `~/.config/local-webdav/config.toml` 创建一个配置文件。

2.  **编辑配置**: 打开生成的 TOML 文件，在 `[[shares]]` 下添加你的文件夹。
    ```toml
    [server]
    address = "127.0.0.1:43621"

    [[shares]]
    name = "docs"
    path = "~/Documents"
    username = "admin"
    password = "yourpassword"

    [[shares]]
    name = "media"
    path = "/Volumes/External/Media"
    # 不设置用户名/密码则无需认证
    ```

3.  **启动服务**: 再次运行启动命令。
    ```bash
    ./local-webdav serve
    ```

4.  **访问**: 使用 WebDAV 客户端（如 Finder、资源管理器等）连接：
    - `http://127.0.0.1:43621/docs/`
    - `http://127.0.0.1:43621/docs/`
    - `http://127.0.0.1:43621/media/`

### Chrome 浏览器扩展 (Cookie Monitor)

本项目包含一个配套的 Chrome 扩展，可以将指定网站的 Cookie 自动同步到你的 WebDAV 服务器。

#### 功能特性
- **自动同步**: 自动检测 Cookie 变化并上传到 WebDAV。
- **Netscape 格式**: 以标准 Netscape 格式导出 Cookie（兼容 `curl`、`wget` 等工具）。
- **域名过滤**: 仅同步你指定的域名的 Cookie。
- **安全可靠**: 认证信息仅保存在本地扩展中。

#### 编译与安装
1.  进入 `extension` 目录:
    ```bash
    cd extension
    pnpm install
    ```
2.  编译扩展:
    ```bash
    pnpm build
    ```
3.  在 Chrome 中加载:
    - 打开 `chrome://extensions/`
    - 开启 “开发者模式”
    - 点击 “加载已解压的扩展程序”，选择 `extension/.output/chrome-mv3` 目录。
