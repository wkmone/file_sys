import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Typography, Row, Col, Spin, Tabs, Empty } from 'antd'
import { FileOutlined } from '@ant-design/icons'
import { filePermissionApi } from '../api/fileApi'
import { folderPermissionApi } from '../api/folderApi'
import FileIcon from '../components/common/FileIcon'
import type { FileItem } from '../types/file'
import type { Folder } from '../types/folder'
import { colors, OFFICE_EDITABLE_EXTS } from '../theme'

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export default function SharedWithMePage() {
  const navigate = useNavigate()
  const [files, setFiles] = useState<FileItem[]>([])
  const [folders, setFolders] = useState<Folder[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    Promise.all([
      filePermissionApi.sharedWithMe().then((res) => setFiles(res.data.data || [])).catch(() => {}),
      folderPermissionApi.sharedWithMe().then((res) => setFolders(res.data.data || [])).catch(() => {}),
    ]).finally(() => setLoading(false))
  }, [])

  const handleFileClick = (file: FileItem) => {
    const isOffice = OFFICE_EDITABLE_EXTS.includes(file.file_ext)
    if (isOffice) {
      window.open(`/editor/${file.id}`, '_blank')
    } else {
      const token = localStorage.getItem('fs_access_token')
      const sep = token ? `?token=${encodeURIComponent(token)}` : ''
      window.open(`/api/v1/files/${file.id}/download${sep}`, '_blank')
    }
  }

  const permLabel = (p?: string) => {
    if (p === 'admin') return '管理'
    if (p === 'write') return '可写'
    return '可读'
  }

  const filesTab = (
    <div>
      {files.length === 0 ? (
        <Empty description="暂无共享文件" style={{ padding: 40 }} />
      ) : (
        <Row gutter={[12, 12]}>
          {files.map((file) => (
            <Col key={file.id} xs={24} sm={12} lg={8} xl={6}>
              <div
                onClick={() => handleFileClick(file)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '12px 16px',
                  border: `1px solid ${colors.border}`,
                  borderRadius: 10,
                  cursor: 'pointer',
                  background: colors.white,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = colors.primary
                  e.currentTarget.style.background = colors.primaryLight
                  e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,82,217,0.1)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = colors.border
                  e.currentTarget.style.background = colors.white
                  e.currentTarget.style.boxShadow = 'none'
                }}
              >
                <FileIcon type={file.file_ext} size={44} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{
                    fontWeight: 500,
                    fontSize: 14,
                    color: colors.textPrimary,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}>
                    {file.name}
                  </div>
                  <div style={{ fontSize: 12, color: colors.textTertiary, marginTop: 2 }}>
                    {formatSize(file.file_size)}
                    {file.shared_by && <span> · 来自 {file.shared_by}</span>}
                    {file.permission && <span> · {permLabel(file.permission)}</span>}
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      )}
    </div>
  )

  const foldersTab = (
    <div>
      {folders.length === 0 ? (
        <Empty description="暂无共享文件夹" style={{ padding: 40 }} />
      ) : (
        <Row gutter={[12, 12]}>
          {folders.map((folder) => (
            <Col key={folder.id} xs={24} sm={12} lg={8} xl={6}>
              <div
                onClick={() => navigate(`/my/files/${folder.id}`)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '12px 16px',
                  border: `1px solid ${colors.border}`,
                  borderRadius: 10,
                  cursor: 'pointer',
                  background: colors.white,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = colors.primary
                  e.currentTarget.style.background = colors.primaryLight
                  e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,82,217,0.1)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = colors.border
                  e.currentTarget.style.background = colors.white
                  e.currentTarget.style.boxShadow = 'none'
                }}
              >
                <FileIcon type="folder" size={44} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{
                    fontWeight: 500,
                    fontSize: 14,
                    color: colors.textPrimary,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}>
                    {folder.name}
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      )}
    </div>
  )

  return (
    <div>
      <Typography.Title level={3} style={{ marginTop: 0, marginBottom: 20, color: colors.textPrimary }}>
        与我共享
      </Typography.Title>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 60 }}><Spin /></div>
      ) : files.length === 0 && folders.length === 0 ? (
        <div style={{
          padding: 60,
          textAlign: 'center',
          border: `1px dashed ${colors.border}`,
          borderRadius: 12,
        }}>
          <FileOutlined style={{ fontSize: 48, color: colors.textTertiary, marginBottom: 16, display: 'block' }} />
          <div style={{ color: colors.textSecondary, fontSize: 15, marginBottom: 4 }}>暂无共享内容</div>
          <div style={{ color: colors.textTertiary, fontSize: 13 }}>其他用户共享给你的文件或文件夹将会显示在这里</div>
        </div>
      ) : (
        <Tabs
          items={[
            { key: 'files', label: `文件 (${files.length})`, children: filesTab },
            { key: 'folders', label: `文件夹 (${folders.length})`, children: foldersTab },
          ]}
        />
      )}
    </div>
  )
}
