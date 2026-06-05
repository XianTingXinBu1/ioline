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

### 说明

- 当前实现为同步设置
- 成功返回后，前端可以立即调用 `GET /api/files/list?path=.`

---

## 4.3 清除当前工作区

### 请求

```http
DELETE /api/workspace/current
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "isSet": false
  }
}
```

---

## 4.4 获取工作区候选目录

### 请求

```http
GET /api/workspaces/candidates
```

### 成功响应示例

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "name": "ioline",
        "path": "/data/data/com.termux/files/home/project/ioline",
        "exists": true,
        "source": "current"
      },
      {
        "name": "project",
        "path": "/data/data/com.termux/files/home/project",
        "exists": true,
        "source": "suggested"
      }
    ]
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 展示名称 |
| `path` | string | 目录绝对路径 |
| `exists` | boolean | 当前实现仅返回存在目录，因此固定为 `true` |
| `source` | string | 候选来源，如 `current` / `suggested` / `default` |

### 当前候选来源

- 当前已设置工作区（若有）
- `$HOME/project`
- `$HOME/projects`
- `$HOME/workspace`
- `$HOME`
- 当前进程工作目录

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

### 排序规则

当前后端实现已保证：

1. 目录在前
2. 文件在后
3. 同类型按名称升序排序

### 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `path` | string | 否 | 相对工作区路径，空时等同于 `.` |

---

## 5.2 获取文件或目录元信息

### 请求

```http
GET /api/files/stat?path=go.mod
```

---

## 6. 文件内容接口

## 6.1 读取文本文件内容

### 请求

```http
GET /api/file/content?path=go.mod
```

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

---

## 7. 文件操作接口

## 7.1 创建文件

### 请求

```http
POST /api/files
Content-Type: application/json
```

---

## 7.2 删除文件或目录

### 请求

```http
DELETE /api/files
Content-Type: application/json
```

---

## 7.3 创建目录

### 请求

```http
POST /api/directories
Content-Type: application/json
```

---

## 7.4 重命名或移动文件/目录

### 请求

```http
PATCH /api/files/move
Content-Type: application/json
```

---

## 8. 终端接口

终端接口依赖工作区，调用前必须先设置当前工作区。

## 8.1 获取终端会话列表

### 请求

```http
GET /api/terminals
```

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

---

## 8.3 调整终端尺寸

### 请求

```http
POST /api/terminals/{id}/resize
Content-Type: application/json
```

---

## 8.4 关闭终端会话

### 请求

```http
DELETE /api/terminals/{id}
```

---

## 8.5 终端 WebSocket 流

### 连接地址

```txt
ws://127.0.0.1:8080/api/terminals/{id}/stream
```

### 当前消息格式

当前实现为原始文本透传：

- 客户端发送：文本消息或二进制消息，内容会直接写入 PTY
- 服务端返回：文本消息，内容为终端输出

---

## 9. 前端接入建议

## 9.1 推荐初始化顺序

前端启动后建议按这个顺序接入：

1. `GET /api/healthz`
2. `GET /api/system/info`
3. `GET /api/workspace/current`
4. 若未设置工作区，调用 `GET /api/workspaces/candidates`
5. 用户选择后调用 `PUT /api/workspace/current`
6. `GET /api/files/list?path=.`
7. 根据文件树继续请求 `list`、`stat`、`content`

---

## 10. 当前未实现的接口范围

以下能力目前尚未提供独立 API：

- 搜索 API
- Git API
- 文件变更监听
- 设置/偏好配置
- LSP/代码智能
- 最近工作区持久化
- 多工作区
- 鉴权
