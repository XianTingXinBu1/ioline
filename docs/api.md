# ioline Backend API 文档

本文档描述当前 ioline 后端已实现的接口，供前端接入与联调用作参考。

## 1. 基础说明

### 1.1 默认服务地址

本地开发默认监听：

```txt
http://127.0.0.1:8080
```

### 1.2 统一响应格式

#### 成功响应

```json
{
  "success": true,
  "data": {}
}
```

#### 失败响应

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "error message"
  }
}
```

### 1.3 工作区前置要求

当前后端默认**无工作区**。

多数文件接口和终端接口在调用前，必须先设置工作区：

```http
PUT /api/workspace/current
```

### 1.4 路径规则

除设置工作区外，文件相关接口统一使用：

- **相对工作区路径**
- 不允许传绝对路径
- 不允许越出工作区根目录

示例：

```txt
.
go.mod
internal/server/server.go
tmp/demo.txt
```

---

## 2. 错误码参考

当前接口中常见错误码：

| code | 说明 |
| --- | --- |
| `METHOD_NOT_ALLOWED` | HTTP 方法不允许 |
| `INVALID_JSON` | 请求体 JSON 非法 |
| `INVALID_WORKSPACE` | 工作区路径无效 |
| `WORKSPACE_NOT_CONFIGURED` | 尚未设置工作区 |
| `INVALID_PATH` | 路径非法或越界 |
| `NOT_FOUND` | 文件、目录或资源不存在 |
| `ALREADY_EXISTS` | 目标已存在 |
| `NOT_REGULAR_FILE` | 目标不是普通文件 |
| `UNSUPPORTED_FILE` | 二进制文件或可执行文件不支持文本打开 |
| `FILE_TOO_LARGE` | 文件过大，不允许按文本打开 |
| `DIRECTORY_NOT_EMPTY` | 非空目录未开启递归删除 |
| `TERMINAL_LIMIT_REACHED` | 终端会话数量达到上限 |
| `TERMINAL_NOT_FOUND` | 终端会话不存在 |
| `INVALID_TERMINAL_SIZE` | 终端尺寸参数非法 |
| `INTERNAL_ERROR` | 服务器内部错误 |

---

## 3. 系统接口

## 3.1 健康检查

### 请求

```http
GET /api/healthz
```

### 成功响应

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

---

## 3.2 系统信息

### 请求

```http
GET /api/system/info
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "name": "ioline",
    "goVersion": "go1.26.3",
    "os": "android",
    "arch": "arm64",
    "termux": true,
    "workspaceSet": false,
    "terminalMaxSessions": 4
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 服务名称 |
| `goVersion` | string | Go 版本 |
| `os` | string | 运行操作系统 |
| `arch` | string | CPU 架构 |
| `termux` | boolean | 是否检测到 Termux 环境 |
| `workspaceSet` | boolean | 当前是否已设置工作区 |
| `terminalMaxSessions` | number | 最大终端会话数 |

---

## 4. 工作区接口

## 4.1 获取当前工作区

### 请求

```http
GET /api/workspace/current
```

### 成功响应示例

未设置时：

```json
{
  "success": true,
  "data": {
    "isSet": false
  }
}
```

已设置时：

```json
{
  "success": true,
  "data": {
    "rootPath": "/data/data/com.termux/files/home/project/ioline",
    "name": "ioline",
    "isSet": true,
    "setAt": "2026-06-05T07:24:04.000000000+08:00"
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `rootPath` | string | 工作区绝对路径，未设置时可能为空 |
| `name` | string | 工作区目录名 |
| `isSet` | boolean | 是否已设置工作区 |
| `setAt` | string | 设置时间，RFC3339 风格时间字符串 |

---

## 4.2 设置当前工作区

### 请求

```http
PUT /api/workspace/current
Content-Type: application/json
```

### 请求体

```json
{
  "rootPath": "/data/data/com.termux/files/home/project/ioline"
}
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "rootPath": "/data/data/com.termux/files/home/project/ioline",
    "name": "ioline",
    "isSet": true,
    "setAt": "2026-06-05T07:24:04.000000000+08:00"
  }
}
```

### 失败场景

- 路径不存在
- 路径不是目录
- JSON 非法

---

## 5. 文件树与文件元信息接口

## 5.1 列出目录内容

### 请求

```http
GET /api/files/list?path=.
```

或：

```http
GET /api/files/list?path=internal
```

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | string | 否 | 相对工作区路径，空时等同于 `.` |

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "name": "apps",
        "path": "apps",
        "type": "directory",
        "size": 4096,
        "modifiedAt": "2026-06-05T07:00:00+08:00",
        "readonly": false,
        "hidden": false
      },
      {
        "name": "go.mod",
        "path": "go.mod",
        "type": "file",
        "size": 89,
        "modifiedAt": "2026-06-05T07:00:00+08:00",
        "readonly": false,
        "hidden": false
      }
    ]
  }
}
```

### `items` 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 当前项名称 |
| `path` | string | 相对工作区路径 |
| `type` | string | `file` 或 `directory` |
| `size` | number | 文件大小，目录值由系统返回 |
| `modifiedAt` | string | 最后修改时间 |
| `readonly` | boolean | 是否只读 |
| `hidden` | boolean | 是否隐藏文件/目录 |

---

## 5.2 获取文件或目录元信息

### 请求

```http
GET /api/files/stat?path=go.mod
```

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | string | 否 | 相对工作区路径，空时等同于 `.` |

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "name": "go.mod",
    "path": "go.mod",
    "type": "file",
    "size": 89,
    "modifiedAt": "2026-06-05T07:00:00+08:00",
    "readonly": false,
    "hidden": false
  }
}
```

---

## 6. 文件内容接口

## 6.1 读取文本文件内容

### 请求

```http
GET /api/file/content?path=go.mod
```

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | string | 是 | 相对工作区路径 |

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "path": "go.mod",
    "content": "module ioline\n\ngo 1.26.3\n",
    "size": 89,
    "modifiedAt": "2026-06-05T07:00:00+08:00",
    "readonly": false,
    "binary": false,
    "lineEnding": "lf"
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `path` | string | 相对工作区路径 |
| `content` | string | 文本内容 |
| `size` | number | 文件大小 |
| `modifiedAt` | string | 最后修改时间 |
| `readonly` | boolean | 是否只读 |
| `binary` | boolean | 当前固定为 `false`，二进制文件会直接报错 |
| `lineEnding` | string | `lf` 或 `crlf` |

### 限制与保护

以下情况会失败：

- 不是普通文件
- 大于当前文本读取限制（约 1 MiB）
- 文件为二进制文件
- 文件具有执行位，按可执行/非文本文件处理

---

## 6.2 保存文本文件内容

### 请求

```http
PUT /api/file/content
Content-Type: application/json
```

### 请求体

```json
{
  "path": "tmp/demo.txt",
  "content": "hello ioline"
}
```

### 规则

- 会自动创建父目录
- 若文件不存在，会直接创建
- 只允许工作区内相对路径

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "path": "tmp/demo.txt",
    "size": 13,
    "modifiedAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 7. 文件操作接口

## 7.1 创建文件

### 请求

```http
POST /api/files
Content-Type: application/json
```

### 请求体

```json
{
  "path": "docs/readme.md",
  "content": "# hello"
}
```

### 规则

- 自动创建父目录
- 若目标已存在，返回 `ALREADY_EXISTS`

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "path": "docs/readme.md",
    "type": "file",
    "modifiedAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 7.2 删除文件或目录

### 请求

```http
DELETE /api/files
Content-Type: application/json
```

### 请求体

删除文件：

```json
{
  "path": "tmp/demo.txt",
  "recursive": false
}
```

递归删除目录：

```json
{
  "path": "tmp/demo-dir",
  "recursive": true
}
```

### 规则

- 文件可直接删除
- 空目录可删除
- 非空目录默认不允许删除
- 非空目录需显式传 `recursive: true`

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "path": "tmp/demo.txt",
    "type": "file",
    "modifiedAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 7.3 创建目录

### 请求

```http
POST /api/directories
Content-Type: application/json
```

### 请求体

```json
{
  "path": "internal/git"
}
```

### 规则

- 自动创建父目录
- 若目标已存在，返回 `ALREADY_EXISTS`

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "path": "internal/git",
    "type": "directory",
    "modifiedAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 7.4 重命名或移动文件/目录

### 请求

```http
PATCH /api/files/move
Content-Type: application/json
```

### 请求体

```json
{
  "fromPath": "docs/a.md",
  "toPath": "docs/b.md"
}
```

### 规则

- 支持文件与目录
- 自动创建目标父目录
- 目标已存在时报错
- 不允许越出工作区
- 源路径和目标路径相同会报错

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "fromPath": "docs/a.md",
    "toPath": "docs/b.md",
    "type": "file",
    "modifiedAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 8. 终端接口

终端接口依赖工作区，调用前必须先设置当前工作区。

## 8.1 获取终端会话列表

### 请求

```http
GET /api/terminals
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "20260605072404-ioline-b",
        "cwd": "/data/data/com.termux/files/home/project/ioline",
        "shell": "/data/data/com.termux/files/usr/bin/bash",
        "status": "running",
        "createdAt": "2026-06-05T07:24:04+08:00"
      }
    ]
  }
}
```

### 会话字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 终端会话 ID |
| `cwd` | string | 当前工作目录 |
| `shell` | string | 启动 shell |
| `status` | string | 当前状态，如 `running` |
| `createdAt` | string | 创建时间 |

---

## 8.2 创建终端会话

### 请求

```http
POST /api/terminals
Content-Type: application/json
```

### 请求体

```json
{
  "cols": 80,
  "rows": 24
}
```

### 规则

- 必须先设置工作区
- `cwd` 为工作区根目录
- shell 优先使用 `$SHELL`
- `$SHELL` 为空时回退 `sh`
- 最多允许 4 个终端会话

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "id": "20260605072404-ioline-b",
    "cwd": "/data/data/com.termux/files/home/project/ioline",
    "shell": "/data/data/com.termux/files/usr/bin/bash",
    "status": "running",
    "createdAt": "2026-06-05T07:24:04+08:00"
  }
}
```

---

## 8.3 调整终端尺寸

### 请求

```http
POST /api/terminals/{id}/resize
Content-Type: application/json
```

### 路径参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 终端会话 ID |

### 请求体

```json
{
  "cols": 100,
  "rows": 30
}
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "id": "20260605072404-ioline-b",
    "cols": 100,
    "rows": 30
  }
}
```

---

## 8.4 关闭终端会话

### 请求

```http
DELETE /api/terminals/{id}
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "id": "20260605072404-ioline-b",
    "status": "closed"
  }
}
```

---

## 8.5 终端 WebSocket 流

### 连接地址

```txt
ws://127.0.0.1:8080/api/terminals/{id}/stream
```

### 使用方式

1. 先调用 `POST /api/terminals` 创建终端会话
2. 获取返回的 `id`
3. 使用该 `id` 建立 WebSocket 连接
4. 前端把用户输入写入 WebSocket
5. 后端把终端输出通过 WebSocket 回推给前端

### 当前消息格式

当前实现为**原始文本透传**：

- 客户端发送：文本消息或二进制消息，内容会直接写入 PTY
- 服务端返回：文本消息，内容为终端输出

### 发送示例

发送：

```txt
pwd
```

或：

```txt
echo hello
```

注意：
- 通常需要自行补 `\n`
- 服务端当前未定义复杂 JSON 协议
- 前端应自行维护输入缓冲、按键行为和特殊键编码

---

## 9. 前端接入建议

## 9.1 推荐初始化顺序

前端启动后建议按这个顺序接入：

1. `GET /api/healthz`
2. `GET /api/system/info`
3. `GET /api/workspace/current`
4. 若未设置工作区，提示用户选择目录
5. `PUT /api/workspace/current`
6. `GET /api/files/list?path=.`
7. 根据文件树继续请求 `list`、`stat`、`content`

---

## 9.2 文件树接入建议

建议采用懒加载：

- 首次只请求根目录
- 用户展开目录节点时，再请求对应 `path`

不建议前端假设后端一次返回整个目录树。

---

## 9.3 文件编辑接入建议

推荐流程：

1. 用户点击文件
2. `GET /api/file/content?path=...`
3. 编辑后 `PUT /api/file/content`
4. 如有必要，再调用 `GET /api/files/stat` 刷新元信息

---

## 9.4 终端接入建议

推荐流程：

1. `POST /api/terminals`
2. 连接 `WS /api/terminals/{id}/stream`
3. 将键盘输入写入 WS
4. 将 WS 输出渲染到终端组件
5. 组件尺寸变化时调用 `POST /api/terminals/{id}/resize`
6. 页面关闭或标签关闭时调用 `DELETE /api/terminals/{id}`

---

## 9.5 错误处理建议

前端建议统一根据：

- HTTP 状态码
- `success`
- `error.code`

进行提示与处理。

例如：

- `WORKSPACE_NOT_CONFIGURED`：提示先选择工作区
- `INVALID_PATH`：提示路径非法
- `ALREADY_EXISTS`：提示目标已存在
- `UNSUPPORTED_FILE`：提示该文件不支持文本打开
- `FILE_TOO_LARGE`：提示文件过大，避免直接打开
- `TERMINAL_LIMIT_REACHED`：提示终端会话数已达上限

---

## 10. 当前未实现的接口范围

以下能力目前尚未提供独立 API：

- 搜索 API
- Git API
- 文件变更监听
- 设置/偏好配置
- LSP/代码智能
- 最近工作区
- 多工作区
- 鉴权

后续接入前端时，如涉及这些功能，需要再补后端实现。
