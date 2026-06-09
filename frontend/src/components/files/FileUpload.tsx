import { useState, useCallback } from 'react'
import { Modal, Upload, message, Progress, Tree, Result, Space, Tag, Typography } from 'antd'
import { InboxOutlined, FolderOpenOutlined, FileOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import type { DataNode } from 'antd/es/tree'
import type { AxiosProgressEvent } from 'axios'
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
          if (parent.children && !parent.children.find((c: DataNode) => c.key === currentPath)) {
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
      const res = await fileApi.batchUpload(formData, (progressEvent: AxiosProgressEvent) => {
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
