# Folder Upload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support uploading entire folders via drag-and-drop into the existing FileUpload dialog, preserving directory structure.

**Architecture:** New `POST /api/v1/files/batch` backend endpoint accepts a JSON manifest describing the folder tree plus file binaries. Backend creates folders as needed (find-or-create by name+parent), then uploads each file. Frontend detects folder drops via `webkitGetAsEntry`, shows a tree preview, calls the batch API, and displays a progress bar with a success/failure summary.

**Tech Stack:** Go + Gin + pgx (backend), React + TypeScript + Ant Design + react-dropzone (frontend)

---

### Task 1: Backend DTOs

**Files:**
- Modify: `D:\code\file_sys\backend\internal\dto\dto.go`

Add batch upload DTOs after the existing `UpdateFileRequest` block (after line 56).

- [ ] **Step 1: Add batch upload DTOs**

Insert the following after line 56 (`type UpdateFileRequest struct { ... }`):

```go
// Batch upload
type BatchUploadManifestEntry struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	IsDirectory bool   `json:"is_directory"`
}

type BatchUploadFileResult struct {
	Path   string `json:"path"`
	FileID string `json:"file_id,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type BatchUploadResponse struct {
	Total     int                     `json:"total"`
	Succeeded int                     `json:"succeeded"`
	Failed    int                     `json:"failed"`
	Results   []BatchUploadFileResult `json:"results"`
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\code\file_sys\backend && go build ./...
```

Expected: builds without errors.

---

### Task 2: FolderRepo FindByNameAndParent

**Files:**
- Modify: `D:\code\file_sys\backend\internal\repository\folder_repo.go`

Add a method to find an existing folder by name and parent, scoped by owner (personal) or team.

- [ ] **Step 1: Add FindByNameAndParent method**

Insert after the `Create` method (after line 28):

```go
func (r *FolderRepo) FindByNameAndParent(ctx context.Context, name string, parentID *string, ownerID string, teamID *string) (*model.Folder, error) {
	f := &model.Folder{}
	var err error
	if teamID != nil && *teamID != "" {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
			 FROM folders WHERE name = $1 AND (parent_id IS NOT DISTINCT FROM $2) AND team_id = $3 AND is_deleted = false`,
			name, parentID, *teamID,
		).Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx,
			`SELECT id, name, parent_id, owner_id, team_id, folder_path, is_deleted, deleted_at, created_at, updated_at
			 FROM folders WHERE name = $1 AND (parent_id IS NOT DISTINCT FROM $2) AND owner_id = $3 AND team_id IS NULL AND is_deleted = false`,
			name, parentID, ownerID,
		).Scan(&f.ID, &f.Name, &f.ParentID, &f.OwnerID, &f.TeamID,
			&f.FolderPath, &f.IsDeleted, &f.DeletedAt, &f.CreatedAt, &f.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd D:\code\file_sys\backend && go build ./...
```

Expected: builds without errors.

---

### Task 3: FileService BatchUpload

**Files:**
- Modify: `D:\code\file_sys\backend\internal\service\file_service.go`

Add the `BatchUpload` method. Requires imports: `"encoding/json"`, `"mime/multipart"`, `"sort"`.

- [ ] **Step 1: Update imports**

Replace the import block (lines 3-15) with:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/repository"
	"file_sys/backend/internal/storage"
)
```

- [ ] **Step 2: Add BatchUpload method**

Insert after the `Upload` method (after line 117):

```go
func (s *FileService) BatchUpload(ctx context.Context, files []*multipart.FileHeader, manifestJSON string, folderID string, ownerID string, teamID *string) (*dto.BatchUploadResponse, error) {
	var manifest []dto.BatchUploadManifestEntry
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}

	// Collect unique directory paths from all entries
	dirSet := make(map[string]bool)
	for _, entry := range manifest {
		dir := filepath.Dir(entry.Path)
		if dir != "." {
			parts := strings.Split(dir, "/")
			for i := range parts {
				dirSet[strings.Join(parts[:i+1], "/")] = true
			}
		}
		if entry.IsDirectory {
			dirSet[entry.Path] = true
		}
	}

	// Sort by depth so parents are created before children
	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	// Create folder hierarchy: path -> folderID
	folderCache := make(map[string]string)
	for _, dir := range dirs {
		parentPath := filepath.Dir(dir)
		var parentID *string
		if parentPath != "." {
			pid := folderCache[parentPath]
			parentID = &pid
		} else if folderID != "" {
			parentID = &folderID
		}

		// Try to find existing folder
		existing, err := s.folderRepo.FindByNameAndParent(ctx, filepath.Base(dir), parentID, ownerID, teamID)
		if err == nil {
			folderCache[dir] = existing.ID
			continue
		}

		// Create new folder
		folder := &model.Folder{
			Name:     filepath.Base(dir),
			ParentID: parentID,
			OwnerID:  ownerID,
			TeamID:   teamID,
		}
		if err := s.folderRepo.Create(ctx, folder); err != nil {
			return nil, fmt.Errorf("create folder %s: %w", dir, err)
		}
		folderCache[dir] = folder.ID
	}

	// Upload files, matching by position (non-directory entries only)
	results := make([]dto.BatchUploadFileResult, 0, len(manifest))
	succeeded, failed := 0, 0
	fileIdx := 0

	for _, entry := range manifest {
		if entry.IsDirectory {
			results = append(results, dto.BatchUploadFileResult{
				Path:   entry.Path,
				Status: "success",
			})
			succeeded++
			continue
		}

		result := dto.BatchUploadFileResult{Path: entry.Path}

		// Determine target folder
		dir := filepath.Dir(entry.Path)
		var targetFolderID string
		if dir != "." {
			targetFolderID = folderCache[dir]
		} else {
			targetFolderID = folderID
		}

		if fileIdx >= len(files) {
			result.Status = "failed"
			result.Error = "file missing in request"
			results = append(results, result)
			failed++
			continue
		}

		fh := files[fileIdx]
		fileIdx++

		reader, err := fh.Open()
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			failed++
			continue
		}

		uploadResp, err := s.Upload(ctx, reader, filepath.Base(entry.Path), targetFolderID, ownerID, teamID)
		reader.(io.Closer).Close()
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			failed++
		} else {
			result.Status = "success"
			result.FileID = uploadResp.ID
			succeeded++
		}
		results = append(results, result)
	}

	return &dto.BatchUploadResponse{
		Total:     len(manifest),
		Succeeded: succeeded,
		Failed:    failed,
		Results:   results,
	}, nil
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd D:\code\file_sys\backend && go build ./...
```

Expected: builds without errors.

---

### Task 4: FileHandler BatchUpload + Route

**Files:**
- Modify: `D:\code\file_sys\backend\internal\handler\file_handler.go`
- Modify: `D:\code\file_sys\backend\internal\router\router.go`

- [ ] **Step 1: Add handler imports**

In `file_handler.go`, add `"encoding/json"` to the import block (line 4-6 area):

```go
import (
	"encoding/json"
	"strconv"

	"file_sys/backend/internal/dto"
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/model"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)
```

- [ ] **Step 2: Add BatchUpload handler**

Insert after the `Upload` handler (after line 78):

```go
func (h *FileHandler) BatchUpload(c *gin.Context) {
	// Parse multipart form with 100MB limit
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil {
		util.ValidationError(c, "request too large")
		return
	}

	manifestStr := c.PostForm("manifest")
	if manifestStr == "" {
		util.ValidationError(c, "missing manifest")
		return
	}

	var manifest []dto.BatchUploadManifestEntry
	if err := json.Unmarshal([]byte(manifestStr), &manifest); err != nil {
		util.ValidationError(c, "invalid manifest JSON: "+err.Error())
		return
	}

	var fileHeaders []*multipart.FileHeader
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		fileHeaders = c.Request.MultipartForm.File["files"]
	}

	folderID := c.PostForm("folder_id")
	rawTeamID := c.PostForm("team_id")
	var teamID *string
	if rawTeamID != "" {
		teamID = &rawTeamID
	}

	resp, err := h.fileService.BatchUpload(c.Request.Context(), fileHeaders, manifestStr, folderID, middleware.GetUserID(c), teamID)
	if err != nil {
		util.InternalError(c, "batch upload failed: "+err.Error())
		return
	}
	util.Created(c, resp)
}
```

- [ ] **Step 3: Register route**

In `router.go`, after line 88 (`files.POST("", fileHandler.Upload)`), add:

```go
				files.POST("/batch", fileHandler.BatchUpload)
```

- [ ] **Step 4: Verify compilation**

```bash
cd D:\code\file_sys\backend && go build ./...
```

Expected: builds without errors.

---

### Task 5: Frontend Types

**Files:**
- Modify: `D:\code\file_sys\frontend\src\types\file.ts`

- [ ] **Step 1: Add batch upload types**

Append to the file:

```typescript
export interface BatchUploadManifestEntry {
  path: string
  size: number
  is_directory: boolean
}

export interface BatchUploadFileResult {
  path: string
  file_id?: string
  status: 'success' | 'failed'
  error?: string
}

export interface BatchUploadResponse {
  total: number
  succeeded: number
  failed: number
  results: BatchUploadFileResult[]
}
```

- [ ] **Step 2: Verify type check**

```bash
cd D:\code\file_sys\frontend && npx tsc --noEmit
```

Expected: no new type errors.

---

### Task 6: Frontend API

**Files:**
- Modify: `D:\code\file_sys\frontend\src\api\fileApi.ts`

- [ ] **Step 1: Import batch types and add batchUpload**

Replace the import on line 3 with:

```typescript
import type { FileItem, FileVersion, BatchUploadResponse } from '../types/file'
```

Add after the `upload` entry (after line 12):

```typescript
  batchUpload: (formData: FormData, onProgress?: (e: ProgressEvent) => void) =>
    apiClient.post<ApiResponse<BatchUploadResponse>>('/files/batch', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000,
      onUploadProgress: onProgress,
    }),
```

- [ ] **Step 2: Verify type check**

```bash
cd D:\code\file_sys\frontend && npx tsc --noEmit
```

Expected: no new type errors.

---

### Task 7: FileUpload Component Enhancement

**Files:**
- Modify: `D:\code\file_sys\frontend\src\components\files\FileUpload.tsx`

This is the core change. The component is rewritten to handle both file and folder uploads via a single drag-and-drop area.

- [ ] **Step 1: Rewrite FileUpload component**

Replace the entire file content with:

```typescript
import { useState, useCallback } from 'react'
import { Modal, Upload, message, Progress, Tree, Result, Space, Tag, Typography } from 'antd'
import { InboxOutlined, FolderOpenOutlined, FileOutlined } from '@ant-design/icons'
import type { UploadFile, DataNode } from 'antd'
import { fileApi } from '../../api/fileApi'
import type { BatchUploadResponse } from '../../types/file'

const { Dragger } = Upload
const { Text } = Typography

interface FileUploadProps {
  open: boolean
  folderId: string | null
  teamId?: string | null
  onClose: () => void
  onSuccess: () => void
}

interface FolderEntry {
  file?: File
  relativePath: string
  isDirectory: boolean
  size: number
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '--'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

async function readDirectoryEntries(reader: FileSystemDirectoryReader): Promise<FileSystemEntry[]> {
  const all: FileSystemEntry[] = []
  const readBatch = (): Promise<FileSystemEntry[]> =>
    new Promise((resolve) => {
      reader.readEntries((entries) => {
        if (entries.length === 0) resolve(all)
        else {
          all.push(...entries)
          readBatch().then(resolve)
        }
      })
    })
  return readBatch()
}

async function traverseDirectory(
  dirEntry: FileSystemDirectoryEntry,
  parentPath: string,
  result: FolderEntry[]
): Promise<void> {
  const relativePath = parentPath ? `${parentPath}/${dirEntry.name}` : dirEntry.name
  result.push({ relativePath, isDirectory: true, size: 0 })

  const reader = dirEntry.createReader()
  const entries = await readDirectoryEntries(reader)

  for (const entry of entries) {
    if (entry.isDirectory) {
      await traverseDirectory(entry as FileSystemDirectoryEntry, relativePath, result)
    } else {
      const fileEntry = entry as FileSystemFileEntry
      const file = await new Promise<File>((resolve) => fileEntry.file(resolve))
      result.push({
        file,
        relativePath: `${relativePath}/${fileEntry.name}`,
        isDirectory: false,
        size: file.size,
      })
    }
  }
}

function buildTreeData(entries: FolderEntry[]): DataNode[] {
  const nodeMap: Record<string, DataNode> = {}
  const roots: DataNode[] = []

  for (const entry of entries) {
    const parts = entry.relativePath.split('/')
    let currentPath = ''

    for (let i = 0; i < parts.length; i++) {
      const prevPath = currentPath
      currentPath = currentPath ? `${currentPath}/${parts[i]}` : parts[i]

      if (!nodeMap[currentPath]) {
        const isLeaf = i === parts.length - 1 && !entry.isDirectory
        const node: DataNode = {
          title: isLeaf ? (
            <Space size={4}>
              <FileOutlined style={{ fontSize: 12, color: '#888' }} />
              <span>{parts[i]}</span>
              <Text type="secondary" style={{ fontSize: 11 }}>
                {formatFileSize(entry.size)}
              </Text>
            </Space>
          ) : (
            parts[i]
          ),
          key: currentPath,
          icon: isLeaf ? undefined : <FolderOpenOutlined style={{ color: '#f5a623' }} />,
          children: [],
        }
        nodeMap[currentPath] = node

        if (prevPath && nodeMap[prevPath]) {
          const parent = nodeMap[prevPath]
          if (parent.children && !parent.children.find((c) => c.key === currentPath)) {
            parent.children.push(node)
          }
        } else if (!prevPath) {
          roots.push(node)
        }
      }
    }
  }

  // Clean empty children arrays on leaf nodes
  const cleanTree = (nodes: DataNode[]): DataNode[] =>
    nodes.map((n) => {
      if (n.children && n.children.length > 0) return { ...n, children: cleanTree(n.children) }
      const { children, ...rest } = n
      return rest
    })

  return cleanTree(roots)
}

export default function FileUpload({ open, folderId, teamId, onClose, onSuccess }: FileUploadProps) {
  const [fileList, setFileList] = useState<UploadFile[]>([])
  const [uploading, setUploading] = useState(false)

  // Folder mode state
  const [folderEntries, setFolderEntries] = useState<FolderEntry[]>([])
  const [isFolderMode, setIsFolderMode] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [progressText, setProgressText] = useState('')
  const [uploadResult, setUploadResult] = useState<BatchUploadResponse | null>(null)

  const resetState = useCallback(() => {
    setFileList([])
    setFolderEntries([])
    setIsFolderMode(false)
    setUploadProgress(0)
    setProgressText('')
    setUploadResult(null)
  }, [])

  const handleClose = useCallback(() => {
    if (uploading) return
    resetState()
    onClose()
  }, [uploading, resetState, onClose])

  // Upload regular files (existing flow)
  const handleFileUpload = useCallback(async () => {
    setUploading(true)
    for (const file of fileList) {
      const formData = new FormData()
      formData.append('file', file as any)
      if (folderId) formData.append('folder_id', folderId)
      if (teamId) formData.append('team_id', teamId)
      try {
        await fileApi.upload(formData)
      } catch {
        message.error(`${file.name} 上传失败`)
      }
    }
    setUploading(false)
    message.success('上传完成')
    setFileList([])
    onSuccess()
    onClose()
  }, [fileList, folderId, teamId, onSuccess, onClose])

  // Upload folder (batch API)
  const handleFolderUpload = useCallback(async () => {
    const fileEntries = folderEntries.filter((e) => !e.isDirectory)

    if (fileEntries.length > 500) {
      message.error('一次最多上传 500 个文件，当前选择了 ' + fileEntries.length + ' 个文件，请分批上传')
      return
    }

    const manifest = folderEntries.map((e) => ({
      path: e.relativePath,
      size: e.size,
      is_directory: e.isDirectory,
    }))

    const formData = new FormData()
    formData.append('manifest', JSON.stringify(manifest))
    if (folderId) formData.append('folder_id', folderId)
    if (teamId) formData.append('team_id', teamId)

    for (const entry of fileEntries) {
      formData.append('files', entry.file!)
    }

    setUploading(true)
    setUploadProgress(0)
    setProgressText('正在批量上传 ' + fileEntries.length + ' 个文件...')

    try {
      const res = await fileApi.batchUpload(formData, (progressEvent) => {
        if (progressEvent.total) {
          const pct = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          setUploadProgress(pct)
        }
      })
      setUploadResult(res.data.data)
      setUploadProgress(100)
    } catch {
      message.error('批量上传失败')
    }
    setUploading(false)
  }, [folderEntries, folderId, teamId])

  const handleFinish = useCallback(() => {
    resetState()
    onSuccess()
    onClose()
  }, [resetState, onSuccess, onClose])

  // Native drop handler for folder detection
  const handleNativeDrop = useCallback(
    (e: React.DragEvent) => {
      const items = e.dataTransfer?.items
      if (!items) return

      for (let i = 0; i < items.length; i++) {
        const entry = items[i].webkitGetAsEntry?.()
        if (entry && entry.isDirectory) {
          e.preventDefault()
          e.stopPropagation()

          const result: FolderEntry[] = []
          traverseDirectory(entry as FileSystemDirectoryEntry, '', result).then(() => {
            setFolderEntries(result)
            setIsFolderMode(true)
            setFileList([])
          })
          return
        }
      }
    },
    []
  )

  const fileCount = folderEntries.filter((e) => !e.isDirectory).length
  const totalSize = folderEntries.reduce((sum, e) => sum + e.size, 0)
  const treeData = buildTreeData(folderEntries)

  // Determine modal config based on state
  const isShowingResult = uploadResult !== null
  const isUploading = uploading

  let modalTitle = '上传文件'
  let modalContent: React.ReactNode
  let okText = '开始上传'
  let onOk: () => void
  let okDisabled = false

  if (isShowingResult) {
    modalTitle = '上传结果'
    okText = '完成'
    onOk = handleFinish
    okDisabled = false
    modalContent = (
      <div>
        <Result
          status={uploadResult.failed === 0 ? 'success' : 'warning'}
          title={`成功 ${uploadResult.succeeded} 个，失败 ${uploadResult.failed} 个`}
          style={{ padding: '16px 0' }}
        />
        {uploadResult.failed > 0 && (
          <div
            style={{
              maxHeight: 200,
              overflow: 'auto',
              background: '#fafafa',
              borderRadius: 8,
              padding: '8px 12px',
            }}
          >
            {uploadResult.results
              .filter((r) => r.status === 'failed')
              .map((r) => (
                <div key={r.path} style={{ fontSize: 12, padding: '2px 0' }}>
                  <Text type="danger">{r.path}</Text>
                  <Text type="secondary" style={{ marginLeft: 8 }}>
                    {r.error}
                  </Text>
                </div>
              ))}
          </div>
        )}
      </div>
    )
  } else if (isUploading) {
    okText = '上传中...'
    okDisabled = true
    onOk = () => {}
    modalContent = (
      <div style={{ textAlign: 'center', padding: '24px 0' }}>
        <Progress type="circle" percent={uploadProgress} />
        <div style={{ marginTop: 16 }}>
          <Text type="secondary">{progressText}</Text>
        </div>
      </div>
    )
  } else if (isFolderMode) {
    modalTitle = '上传文件夹'
    okText = `确认上传 (${fileCount} 个文件)`
    onOk = handleFolderUpload
    okDisabled = fileCount === 0
    const dirCount = folderEntries.filter((e) => e.isDirectory).length
    modalContent = (
      <div>
        <div
          style={{
            background: '#f5f6f7',
            borderRadius: 8,
            padding: '8px 16px',
            marginBottom: 12,
            display: 'flex',
            gap: 16,
          }}
        >
          <Space>
            <FolderOpenOutlined style={{ color: '#f5a623', fontSize: 18 }} />
            <Text strong>{dirCount} 个文件夹</Text>
          </Space>
          <Space>
            <FileOutlined style={{ color: '#888', fontSize: 18 }} />
            <Text strong>{fileCount} 个文件</Text>
          </Space>
          <Space>
            <Text type="secondary">总大小 {formatFileSize(totalSize)}</Text>
          </Space>
        </div>
        <div style={{ maxHeight: 280, overflow: 'auto' }}>
          <Tree
            showIcon
            defaultExpandAll
            treeData={treeData}
            style={{ fontSize: 13 }}
          />
        </div>
        <div style={{ marginTop: 12 }}>
          <a
            onClick={() => {
              setFolderEntries([])
              setIsFolderMode(false)
            }}
            style={{ fontSize: 12 }}
          >
            取消，重新选择文件
          </a>
        </div>
      </div>
    )
  } else {
    onOk = handleFileUpload
    okDisabled = fileList.length === 0
    modalContent = (
      <div onDragOver={(e) => e.preventDefault()} onDrop={handleNativeDrop}>
        <Dragger
          multiple
          fileList={fileList}
          beforeUpload={(file) => {
            setFileList((prev) => [...prev, file as UploadFile])
            return false
          }}
          onRemove={(file) => {
            setFileList((prev) => prev.filter((f) => f.uid !== file.uid))
          }}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
          <p className="ant-upload-hint">
            支持 Office 文档、PDF、图片、文本文件
            <br />
            拖拽文件夹可保留目录结构批量上传
          </p>
        </Dragger>
      </div>
    )
  }

  return (
    <Modal
      title={modalTitle}
      open={open}
      onCancel={handleClose}
      onOk={onOk}
      confirmLoading={isUploading && !uploadResult}
      okText={okText}
      okButtonProps={{ disabled: okDisabled }}
      cancelText={isShowingResult ? undefined : '取消'}
      cancelButtonProps={{ style: isShowingResult ? { display: 'none' } : undefined }}
      width={isFolderMode ? 520 : 420}
    >
      {modalContent}
    </Modal>
  )
}
```

- [ ] **Step 2: Verify type check**

```bash
cd D:\code\file_sys\frontend && npx tsc --noEmit
```

Expected: no type errors. Fix any that appear.

- [ ] **Step 3: Verify the app builds**

```bash
cd D:\code\file_sys\frontend && npm run build
```

Expected: build succeeds.

---

### Task 8: Backend Integration Verification

**Files:** No changes. Verify the full flow works end-to-end.

- [ ] **Step 1: Start backend**

```bash
cd D:\code\file_sys\backend && go run ./cmd/server
```

Expected: server starts on port 8080 without errors.

- [ ] **Step 2: Start frontend**

```bash
cd D:\code\file_sys\frontend && npm run dev
```

Expected: dev server starts on port 3000.

- [ ] **Step 3: Manual smoke test**

1. Open the app, log in, navigate to `/my/files`
2. Click "上传文件" to open the dialog
3. Drag a single file in — verify existing file upload still works
4. Drag a folder from the filesystem into the dialog — verify:
   - The folder tree preview appears
   - The file count and size summary are correct
   - Click "确认上传" starts the batch upload
   - Progress spinner shows during upload
   - Results screen shows success/failure counts
   - Clicking "完成" refreshes the file list
5. Verify the uploaded folder structure appears correctly in the file browser
