# ioline 文档索引

## 开发入口

项目根目录现已提供统一 `Makefile` 入口：

```bash
make help
make dev
make build
make check
```

### 常用命令

- `make dev`
  - 同时启动后端开发服务与前端 Vite 开发服务
- `make dev-backend`
  - 仅启动后端
- `make dev-frontend`
  - 仅启动前端
- `make build`
  - 构建后端与前端
- `make check`
  - 进行保守的前后端构建检查

## 其他文档

- `docs/api.md`：后端 API 文档
