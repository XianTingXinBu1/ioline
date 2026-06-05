# 前端接入工作区与文件树的后端诉求说明

本文档从前端接入视角出发，说明当前 ioline 前端在对接工作区、文件树与文件打开流程时，希望后端补充或明确的接口能力与行为约定。

本文档目标不是替代正式 API 文档，而是帮助前后端快速对齐：

- 当前已有接口已经能支持哪些流程
- 当前还缺哪些能力会影响前端体验
- 哪些诉求是必须项，哪些是建议项
- 如果暂时不实现，前端可如何降级

---

## 1. 当前背景

当前后端已经提供：

- `GET /api/healthz`
- `GET /api/system/info`
- `GET /api/workspace/current`
- `PUT /api/workspace/current`
- `DELETE /api/workspace/current`
- `GET /api/workspaces/candidates`
- `GET /api/files/list?path=...`
- `GET /api/files/stat?path=...`
- `GET /api/file/content?path=...`
- `PUT /api/file/content`

这些接口已经足够支持以下能力：

1. 检查服务可用性
2. 查询当前是否已设置工作区
3. 设置工作区
4. 清空当前工作区
5. 在未设置工作区时获取候选目录
6. 在已设置工作区的前提下列出根目录或子目录内容
7. 打开文本文件内容
8. 保存文本文件内容

也就是说，前端已经可以开始接入工作区选择、文件树与文件编辑主流程。

当前主要缺口集中在：

- 若要支持真正的“逐层浏览目录后选为工作区”，还缺目录浏览接口
- 文件树展示时，若希望更细优化，可再考虑 `hasChildren`
- 后续若要提升体验，可考虑最近工作区持久化

---

## 2. 当前前端期望的产品流程

前端希望支持如下流程：

### 2.1 无工作区时

- 初始打开项目时，如果后端返回当前未设置工作区
- 编辑器先显示一个临时文本页（欢迎页/提示页）
- 用户打开侧栏文件树时，不直接显示空白
- 而是显示：
  - 当前未选择工作区
  - 一个“点击以选择工作区”的入口

### 2.2 已设置工作区后

- 前端调用设置工作区接口
- 设置成功后，以该目录作为根目录
- 调用 `GET /api/files/list?path=.` 拉取根目录内容
- 文件树以工作区根目录展开
- 后续目录采用懒加载方式逐层展开

这个流程里，当前最关键的前后端对齐点已具备：

- 前端可以通过 `GET /api/workspaces/candidates` 获取候选目录
- 前端可以在需要时通过 `DELETE /api/workspace/current` 回到无工作区态

---

## 3. 已落地能力

## 3.1 工作区候选列表接口（已实现）

### 接口

```http
GET /api/workspaces/candidates
```

### 当前后端行为

当前实现会返回一组稳定、去重后的存在目录，候选来源包含：

- 当前已设置工作区（若有）
- `$HOME/project`
- `$HOME/projects`
- `$HOME/workspace`
- `$HOME`
- 当前进程工作目录

### 当前响应结构

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
| `exists` | boolean | 当前实现固定为 `true`，仅返回存在目录 |
| `source` | string | 候选来源，如 `current` / `suggested` / `default` |

### 前端使用建议

- 无工作区时调用该接口
- 展示候选目录列表
- 用户点击后调用 `PUT /api/workspace/current`
- 成功后立即请求 `GET /api/files/list?path=.`

---

## 3.2 清除当前工作区接口（已实现）

### 接口

```http
DELETE /api/workspace/current
```

### 响应示例

```json
{
  "success": true,
  "data": {
    "isSet": false
  }
}
```

### 前端使用建议

适用于：

- 返回欢迎页
- 切换项目时先清空状态
- 主动退出当前工作区

---

## 4. P1：仍建议后续补充的能力

## 4.1 目录浏览接口（用于选择工作区）

如果希望前端支持真正的“点选目录”而不是点候选项，建议后续补充一个专门用于浏览目录的接口。

### 建议接口

```http
GET /api/workspaces/browse?path=/data/data/com.termux/files/home
```

也可以命名为：

```http
GET /api/directories/browse?path=/data/data/com.termux/files/home
```

### 建议响应结构

```json
{
  "success": true,
  "data": {
    "currentPath": "/data/data/com.termux/files/home",
    "parentPath": "/data/data/com.termux/files",
    "items": [
      {
        "name": "project",
        "path": "/data/data/com.termux/files/home/project",
        "type": "directory",
        "readonly": false
      }
    ]
  }
}
```

### 说明

这个接口当前仍未实现，但已不是前端启动接入的阻塞项。

---

## 5. 已明确的行为约定

## 5.1 `GET /api/files/list` 的排序规则（已明确）

当前后端实现已保证：

1. 目录在前
2. 文件在后
3. 同类型按名称升序排序

前端可以基于这一行为组织文件树，如需额外排序，也可自行处理。

---

## 5.2 `GET /api/files/list` 的 `hasChildren`（未实现，可选）

当前未返回 `hasChildren`。

前端可采用保守策略：

- 目录默认都允许点开
- 点开后若返回空列表，则视为空目录

该项优先级较低。

---

## 5.3 `PUT /api/workspace/current` 成功后是否可立即使用工作区（已明确）

当前后端实现为**同步设置工作区**，因此：

1. `PUT /api/workspace/current` 成功返回
2. 随后立刻调用 `GET /api/files/list?path=.`
3. 可以立即拿到新工作区根目录内容

前端可直接按串行逻辑处理，不需要额外等待。

---

## 5.4 错误码稳定性（已明确）

以下错误码目前已作为稳定约定对待，前端可基于它们做交互分流：

- `WORKSPACE_NOT_CONFIGURED`
- `INVALID_WORKSPACE`
- `INVALID_PATH`
- `NOT_FOUND`
- `UNSUPPORTED_FILE`
- `FILE_TOO_LARGE`
- `ALREADY_EXISTS`
- `DIRECTORY_NOT_EMPTY`

---

## 6. P2：可选增强能力

## 6.1 最近工作区持久化（可选）

当前 `GET /api/workspaces/candidates` 还没有做真正的最近工作区持久化。

如果未来需要：

- 可把最近成功设置的工作区记录到本地配置
- 候选列表可优先展示最近使用记录

这会明显改善重复打开项目的体验。

---

## 7. 当前前端可采用的推荐方案

在当前后端能力下，推荐前端按以下方式接入：

### 无工作区时

1. `GET /api/workspace/current`
2. 若 `isSet=false`，显示欢迎页和“选择工作区”入口
3. 调用 `GET /api/workspaces/candidates`
4. 展示候选目录列表
5. 用户选择目录后，调用 `PUT /api/workspace/current`
6. 成功后调用 `GET /api/files/list?path=.`

### 已有工作区时

1. `GET /api/workspace/current`
2. 直接调用 `GET /api/files/list?path=.``
3. 懒加载目录
4. 点击文件时调用 `GET /api/file/content?path=...`
5. 编辑后调用 `PUT /api/file/content`

### 清空工作区时

1. `DELETE /api/workspace/current`
2. 回到欢迎页和候选工作区列表

---

## 8. 建议的后端后续实现优先级

### 当前已完成

- `GET /api/workspaces/candidates`
- `DELETE /api/workspace/current`
- 明确 `GET /api/files/list` 排序规则
- 明确 `PUT /api/workspace/current` 成功后可立即使用
- 稳定错误码命名

### 下一步建议（P1）

- `GET /api/workspaces/browse?path=...`
- 可选的 `items[].hasChildren`
- 最近工作区持久化

---

## 9. 总结

从前端接入角度看：

- 当前后端 API 已经足够支持“无工作区 -> 选择候选工作区 -> 打开根目录文件树 -> 打开文件 -> 保存文件”的主流程
- 当前最主要的体验问题已经从“完全不会选工作区”下降为“是否需要更完整的目录浏览器”
- 如果后续再补目录浏览接口，前端就可以进一步做出更自然的工作区目录选择器

在当前阶段，推荐推进策略是：

1. 前端先接 `GET /api/workspaces/candidates`
2. 完成无工作区态、工作区设置、根目录文件树接入
3. 之后再逐步补目录展开、文件打开、保存与终端能力
4. 目录浏览器作为下一轮增强能力推进
