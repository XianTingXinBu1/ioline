# 后端第一阶段最小压测说明

本文档描述 ioline 后端当前第一阶段最小压测脚本的使用方式。

## 1. 目标

当前压测脚本用于提供可重复、低依赖的手动压力验证能力，优先覆盖：

- `GET /api/files/list`
- `POST /api/search/text`

当前方案不追求完整性能评估，只提供基础负载观察入口。

## 2. 脚本位置

```txt
scripts/stress/generate_workspace_fixture.sh
scripts/stress/test_files_list.sh
scripts/stress/test_search_text.sh
```

## 3. 环境前提

- 后端开发服务已启动
- 默认服务地址为 `http://127.0.0.1:9650`
- 当前设备允许在本地生成测试目录和文件

## 4. 生成固定 fixture

```bash
bash scripts/stress/generate_workspace_fixture.sh .tmp/stress-workspace
```

可选环境变量：

- `DIR_COUNT`：目录数量，默认 `12`
- `FILES_PER_DIR`：每个目录的文件数，默认 `15`
- `LINES_PER_FILE`：每个文件的行数，默认 `40`
- `LARGE_FILES`：额外较大文本文件数量，默认 `4`

生成内容包括：

- 多层模块目录
- 文本文件
- 少量较大文本文件
- `.git` / `node_modules` / `.runtime` / `.tmp` 等忽略目录样例

## 5. 文件列表压测

```bash
WORKSPACE_PATH=.tmp/stress-workspace \
REQUESTS=20 \
CONCURRENCY=4 \
bash scripts/stress/test_files_list.sh
```

可选环境变量：

- `BASE_URL`：默认 `http://127.0.0.1:9650`
- `WORKSPACE_PATH`：待压测工作区路径，必填
- `TARGET_PATH`：列表目录路径，默认 `.`
- `REQUESTS`：总请求数，默认 `20`
- `CONCURRENCY`：并发数，默认 `4`

## 6. 文本搜索压测

```bash
WORKSPACE_PATH=.tmp/stress-workspace \
QUERY=search-keyword \
REQUESTS=12 \
CONCURRENCY=3 \
bash scripts/stress/test_search_text.sh
```

可选环境变量：

- `BASE_URL`：默认 `http://127.0.0.1:9650`
- `WORKSPACE_PATH`：待压测工作区路径，必填
- `QUERY`：搜索关键字，默认 `search-keyword`
- `REQUESTS`：总请求数，默认 `12`
- `CONCURRENCY`：并发数，默认 `3`

## 7. 输出指标

当前脚本输出以下粗略指标：

- `requests`：总请求数
- `concurrency`：并发数
- `success`：HTTP 200 请求数
- `failed`：非 200 或失败请求数
- `totalTime`：整批请求总耗时（秒级粗略值）
- `avgTime`：平均单次请求耗时（毫秒，粗略值）
- `maxTime`：最大单次请求耗时（毫秒，粗略值）

## 8. 当前局限性

当前脚本属于第一阶段最小方案，存在以下局限：

- 仅覆盖 `files/list` 与 `search/text`
- 未覆盖 WebSocket 与终端流
- 未提供 P95 / P99 延迟
- 未统计 CPU / 内存占用
- 未进行长时间 soak test
- 结果受本地设备负载与 Termux 环境影响较大

## 9. 建议使用方式

建议先用较小参数做保守验证，再逐步增大：

1. 先验证 `REQUESTS=10~20`
2. 再逐步增加并发
3. 观察是否出现错误率上升、耗时明显恶化、服务异常退出

如果后续继续扩展，建议下一阶段补充：

- `GET /api/search/files` 压测
- 文件读取与保存压测
- 终端创建/关闭压测
- 资源观测与结果落盘
