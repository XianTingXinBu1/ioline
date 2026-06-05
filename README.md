# ioline

ioline 是一个适用于termux的移动端代码编辑器项目，当前采用：

- **Go** 后端
- **Vue 3 + Vite** 前端

目前处于**开发**阶段


## 项目结构

```txt
apps/server/      Go 后端入口
internal/         后端内部实现（workspace/files/search/terminal/server）
web/              前端应用（Vue 3 + Vite）
scripts/          开发、联调、压测脚本
docs/             项目文档与 API 文档
```

## 快速开始

### 1. 查看可用命令

```bash
make help
```

### 2. 启动开发环境

```bash
make dev
```

默认开发环境会：

- 启动 Go 后端开发服务
- 启动前端 Vite 开发服务
- 后端开发端口默认使用 `9650`

### 3. 查看运行状态

```bash
make status
```

### 4. 停止开发环境

```bash
make stop
```

## 常用命令

```bash
make dev
make stop
make status
make build
make check
make test-backend
make test-backend-smoke
```

说明：

- `make dev`
  - 启动前后端开发服务
- `make stop`
  - 停止统一管理的开发服务
- `make status`
  - 查看当前开发服务状态
- `make build`
  - 构建后端与前端
- `make check`
  - 执行保守构建检查
- `make test-backend`
  - 执行后端单元测试与 handler/API 测试（`go test ./...`）
- `make test-backend-smoke`
  - 对已运行的后端开发服务执行轻量联调脚本 `scripts/test_backend.sh`

## 后端开发说明

当前后端重点模块包括：

- `internal/workspace`
  - 工作区状态、候选目录、目录浏览
- `internal/files`
  - 工作区内文件与目录操作
- `internal/search`
  - 文件名搜索与文本搜索
- `internal/terminal`
  - 基础终端会话管理
- `internal/server`
  - HTTP 路由、handler 与错误映射

后端当前已经建立基础测试护栏：

- service 层单元测试（workspace / files / search）
- `internal/server` handler/API 测试
- GitHub Actions 后端测试工作流
- 本地 smoke 测试脚本

## 压力测试与联调

当前已提供：

- `scripts/test_backend.sh`
  - 后端主流程轻量联调脚本
- `scripts/stress/`
  - 第一阶段最小压测脚本

压测说明见：

- `docs/backend-stress-test.md`

## 文档索引

- `docs/README.md`
  - 文档索引
- `docs/api.md`
  - 后端 API 文档
- `docs/backend-stress-test.md`
  - 后端第一阶段最小压测说明