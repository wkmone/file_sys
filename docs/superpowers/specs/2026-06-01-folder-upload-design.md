# Folder Upload Design

## Overview

支持用户通过拖拽文件夹到上传对话框，批量上传整个目录结构，保持原有层级关系。

## User Flow

```
拖拽文件夹到上传对话框 → 递归解析目录树 → 预览确认 → 批量上传 → 结果汇总
```

## Key Decisions

| 决策点 | 选择 |
|--------|------|
| 后端方案 | 新增批量接口 `POST /api/v1/files/batch`，接收 manifest + files，自动创建目录后上传文件 |
| UI 入口 | 仅拖拽支持：在现有 FileUpload 对话框的 Dragger 区域检测文件夹拖入，不增加独立按钮 |
| 进度展示 | 整体进度条 + "正在上传 X/N 个文件" |
| 错误处理 | 继续上传 + 完成后展示成功/失败汇总 |

## Backend

### New Endpoint

**`POST /api/v1/files/batch`** (protected, multipart/form-data)

| Field | Type | Description |
|-------|------|-------------|
| `files` | file[] | 多个文件二进制（字段名可重复） |
| `manifest` | string (JSON) | 文件清单：`[{"path":"子目录/文档.docx","size":1234}, ...]` |
| `folder_id` | string | 目标父文件夹 UUID（可选） |
| `team_id` | string | 团队 ID（可选） |

### Service Logic

1. 解析 manifest JSON，提取所有唯一目录路径
2. 按层级顺序创建缺失的文件夹：
   - 拆解路径 `"a/b/c"` → `["a", "a/b", "a/b/c"]`
   - 对每层执行 find-or-create：同名+同父 = 复用已有文件夹
3. 按 manifest 顺序处理每个文件：
   - 找到文件所属的文件夹 ID
   - 调用现有单文件上传逻辑（MIME 检测、hash 去重、存储、DB 记录）
   - 失败时收集错误，不中断
4. 返回 `BatchUploadResponse`

### DTO

```go
type BatchUploadManifestEntry struct {
    Path        string `json:"path"`
    Size        int64  `json:"size"`
    IsDirectory bool   `json:"is_directory"`
}

当 `is_directory: true` 时表示空文件夹，只创建目录不关联文件。

type BatchUploadResponse struct {
    Total     int                       `json:"total"`
    Succeeded int                       `json:"succeeded"`
    Failed    int                       `json:"failed"`
    Results   []BatchUploadFileResult   `json:"results"`
}

type BatchUploadFileResult struct {
    Path   string `json:"path"`
    FileID string `json:"file_id,omitempty"`
    Status string `json:"status"` // "success" | "failed"
    Error  string `json:"error,omitempty"`
}
```

### Files Changed

- `internal/handler/file_handler.go` — new `BatchUpload` handler
- `internal/service/file_service.go` — new `BatchUpload` method
- `internal/repository/folder_repo.go` — new `FindByNameAndParent` method
- `internal/dto/dto.go` — new DTOs
- `internal/router/router.go` — new route

## Frontend

### Folder Drop Detection

在 `FileUpload.tsx` 中，对 Dragger 区域添加原生 drop 事件处理：

1. 从 `event.dataTransfer.items` 获取 `DataTransferItem`
2. 调用 `item.webkitGetAsEntry()` 获取 `FileSystemEntry`
3. 若 `isDirectory === true` → 文件夹上传模式
4. 递归遍历 `FileSystemDirectoryEntry`，读取所有文件，构建 `{file, relativePath}` 列表
5. 若拖入的是普通文件 → 走现有文件上传逻辑

### Preview

检测到文件夹拖入后，对话框内容切换为预览模式：
- 展示文件夹树结构（使用 antd Tree 组件，只读）
- 显示总文件数和总大小
- "确认上传" 按钮触发批量上传

### Upload

1. 构建 FormData：manifest JSON 字符串 + 所有文件
2. 调用 `fileApi.batchUpload(formData)`
3. 整体进度：`正在上传 {current}/{total} 个文件` + Progress 进度条
4. 完成后展示汇总：
   - 成功 X 个，失败 Y 个
   - 失败项列表（文件名 + 原因）
5. 用户关闭对话框后刷新文件列表

### Files Changed

- `src/components/files/FileUpload.tsx` — 核心改动
- `src/api/fileApi.ts` — 新增 `batchUpload` 函数
- `src/types/file.ts` — 新增批量上传相关类型

## Edge Cases

| 场景 | 处理 |
|------|------|
| 空文件夹 | manifest 中标注 `isDirectory: true`，后端创建空目录 |
| 同名文件夹 | find-or-create：同父+同名 → 复用已有 ID |
| 超大文件夹（>500 文件） | 前端限制，超出提示分批上传 |
| 不支持的文件类型 | 后端按现有 MIME 逻辑拒绝，标记为 failed |
| 网络中断 | 已上传的文件保留，未上传的标记为 failed |
| 拖入多个文件夹 | 不支持，仅取第一个文件夹 |
