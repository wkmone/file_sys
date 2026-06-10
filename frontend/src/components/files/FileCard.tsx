import { Dropdown, message } from 'antd'
import {
  MoreOutlined,
  EditOutlined,
  DownloadOutlined,
  DeleteOutlined,
  ShareAltOutlined,
  TeamOutlined,
} from '@ant-design/icons'
import type { FileItem } from '../../types/file'
import { fileApi } from '../../api/fileApi'
import FileIcon from '../common/FileIcon'
import { colors, OFFICE_EDITABLE_EXTS } from '../../theme'

interface FileCardProps {
  file: FileItem
  onEdit: (file: FileItem) => void
  onDelete: () => void
  onShare: (file: FileItem) => void
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export default function FileCard({ file, onEdit, onDelete, onShare }: FileCardProps) {
  const handleDownload = () => {
    window.open(fileApi.downloadUrl(file.id), '_blank')
  }

  const handleDelete = async () => {
    try {
      await fileApi.delete(file.id)
      message.success('已移至回收站')
      onDelete()
    } catch {
      message.error('删除失败')
    }
  }

  const menuItems = [
    { key: 'edit', icon: <EditOutlined />, label: '编辑', onClick: () => onEdit(file) },
    { key: 'share', icon: <ShareAltOutlined />, label: '共享', onClick: () => onShare(file) },
    { key: 'download', icon: <DownloadOutlined />, label: '下载', onClick: handleDownload },
    { type: 'divider' as const },
    { key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true, onClick: handleDelete },
  ]

  const isOffice = OFFICE_EDITABLE_EXTS.includes(file.file_ext)

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '20px 12px 16px',
        border: `1px solid ${colors.border}`,
        borderRadius: 12,
        cursor: 'pointer',
        background: colors.white,
        transition: 'all 0.2s',
        position: 'relative',
      }}
      onClick={() => {
        if (isOffice) onEdit(file)
        else handleDownload()
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = colors.primary
        e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,82,217,0.1)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = colors.border
        e.currentTarget.style.boxShadow = 'none'
      }}
    >
      {/* Shared indicator in top-left */}
      {(file.permission || file.shared_by) && (
        <div style={{ position: 'absolute', top: 8, left: 8 }}>
          <TeamOutlined style={{ fontSize: 14, color: colors.primary }} title="已共享" />
        </div>
      )}

      {/* Action menu in top-right */}
      <div style={{ position: 'absolute', top: 8, right: 8 }}>
        <Dropdown menu={{ items: menuItems }} trigger={['click']}>
          <MoreOutlined
            style={{ fontSize: 18, color: colors.textTertiary, cursor: 'pointer', padding: 4 }}
            onClick={(e) => e.stopPropagation()}
          />
        </Dropdown>
      </div>

      <div style={{ marginBottom: 12, marginTop: 8 }}>
        <FileIcon type={file.file_ext} size={56} />
      </div>

      {/* Name */}
      <div style={{
        fontWeight: 500,
        fontSize: 14,
        color: colors.textPrimary,
        textAlign: 'center',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
        width: '100%',
        marginBottom: 6,
      }}>
        {file.name}
      </div>

      {/* Meta */}
      <div style={{ fontSize: 12, color: colors.textTertiary, textAlign: 'center' }}>
        {formatSize(file.file_size)}
        <span style={{ margin: '0 6px' }}>·</span>
        v{file.current_version}
      </div>
    </div>
  )
}
