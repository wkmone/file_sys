import { Dropdown, Table, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useEffect, useRef, useState, useCallback } from 'react'
import {
  MoreOutlined,
  EditOutlined,
  DownloadOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import type { FileItem } from '../../types/file'
import type { Folder } from '../../types/folder'
import { fileApi } from '../../api/fileApi'
import FileIcon from '../common/FileIcon'
import { colors } from '../../theme'

interface FileTableProps {
  files: FileItem[]
  folders: Folder[]
  loading: boolean
  onEnterFolder: (folder: Folder) => void
  onEditFile: (file: FileItem) => void
  onRefresh: () => void
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export default function FileTable({ files, folders, loading, onEnterFolder, onEditFile, onRefresh }: FileTableProps) {
  const handleDownload = (file: FileItem) => {
    window.open(fileApi.downloadUrl(file.id), '_blank')
  }

  const handleDelete = async (file: FileItem) => {
    try {
      await fileApi.delete(file.id)
      message.success('已移至回收站')
      onRefresh()
    } catch {
      message.error('删除失败')
    }
  }

  const isFolder = (r: FileItem | Folder) => 'parent_id' in r
  const folderFirst = (a: FileItem | Folder, b: FileItem | Folder) => {
    if (isFolder(a) && !isFolder(b)) return -1
    if (!isFolder(a) && isFolder(b)) return 1
    return 0
  }

  const columns: ColumnsType<FileItem | Folder> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      sorter: (a, b) => folderFirst(a, b) || a.name.localeCompare(b.name),
      render: (name: string, record) => {
        const isFolder = 'parent_id' in record
        return (
          <div
            style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer' }}
            onClick={() => isFolder ? onEnterFolder(record as Folder) : undefined}
            onDoubleClick={() => {
              if (!isFolder) {
                const f = record as FileItem
                const isOffice = ['.docx', '.xlsx', '.pptx'].includes(f.file_ext)
                if (isOffice) onEditFile(f)
                else handleDownload(f)
              }
            }}
          >
            {isFolder
              ? <FileIcon type="folder" size={26} />
              : <FileIcon type={(record as FileItem).file_ext} size={26} />
            }
            <span style={{ fontWeight: isFolder ? 500 : 400, color: colors.textPrimary, fontSize: 14 }}>
              {name}
            </span>
          </div>
        )
      },
    },
    {
      title: '所有者',
      key: 'owner',
      width: 120,
      render: (_, record) => {
        if ('parent_id' in record) return '-'
        return (record as FileItem).owner_name || '-'
      },
    },
    {
      title: '大小',
      dataIndex: 'file_size',
      key: 'size',
      width: 90,
      sorter: (a, b) => {
        const sizeA = 'file_size' in a ? a.file_size : 0
        const sizeB = 'file_size' in b ? b.file_size : 0
        return folderFirst(a, b) || sizeA - sizeB
      },
      render: (_, record) => {
        if ('parent_id' in record) return '-'
        return formatSize((record as FileItem).file_size)
      },
    },
    {
      title: '类型',
      key: 'type',
      width: 100,
      sorter: (a, b) => {
        const typeA = isFolder(a) ? '' : (a as FileItem).file_ext
        const typeB = isFolder(b) ? '' : (b as FileItem).file_ext
        return folderFirst(a, b) || typeA.localeCompare(typeB)
      },
      render: (_, record) => {
        if ('parent_id' in record) return '文件夹'
        const ext = (record as FileItem).file_ext
        if (!ext) return '-'
        return ext.replace('.', '').toUpperCase() + ' 文件'
      },
    },
    {
      title: '修改时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      sorter: (a, b) => folderFirst(a, b) || new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime(),
      render: (val: string) => {
        const d = new Date(val)
        return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      },
    },
    {
      title: '版本',
      key: 'version',
      width: 60,
      align: 'center',
      render: (_, record) => {
        if ('current_version' in record) return 'v' + (record as FileItem).current_version
        return '-'
      },
    },
    {
      title: '',
      key: 'actions',
      width: 40,
      render: (_, record) => {
        if ('parent_id' in record) return null
        const file = record as FileItem
        const fileMenuItems = [
          { key: 'edit', icon: <EditOutlined />, label: '编辑', onClick: () => onEditFile(file) },
          { key: 'download', icon: <DownloadOutlined />, label: '下载', onClick: () => handleDownload(file) },
          { type: 'divider' as const },
          { key: 'delete', icon: <DeleteOutlined />, label: '删除', danger: true, onClick: () => handleDelete(file) },
        ]
        return (
          <Dropdown menu={{ items: fileMenuItems }} trigger={['click']}>
            <MoreOutlined style={{ fontSize: 18, color: colors.textTertiary, cursor: 'pointer' }} onClick={(e) => e.stopPropagation()} />
          </Dropdown>
        )
      },
    },
  ]

  const dataSource = [
    ...folders.map((f) => ({ ...f, key: `folder-${f.id}` })),
    ...files.map((f) => ({ ...f, key: `file-${f.id}` })),
  ]

  const wrapperRef = useRef<HTMLDivElement>(null)
  const [scrollY, setScrollY] = useState(0)

  const measure = useCallback(() => {
    if (wrapperRef.current) {
      const h = wrapperRef.current.clientHeight - 1 // -1 to avoid double scrollbar
      if (h > 0) setScrollY(h)
    }
  }, [])

  useEffect(() => {
    measure()
    const ro = new ResizeObserver(measure)
    if (wrapperRef.current) ro.observe(wrapperRef.current)
    window.addEventListener('resize', measure)
    return () => { ro.disconnect(); window.removeEventListener('resize', measure) }
  }, [measure])

  // Inject scrollbar styles
  useEffect(() => {
    const id = 'file-table-scrollbar-style'
    if (document.getElementById(id)) return
    const style = document.createElement('style')
    style.id = id
    style.textContent = `
      .file-table-wrapper .ant-table-body {
        scrollbar-width: thin;
        scrollbar-color: transparent transparent;
      }
      .file-table-wrapper .ant-table-body:hover {
        scrollbar-color: ${colors.border} transparent;
      }
      .file-table-wrapper .ant-table-body::-webkit-scrollbar {
        width: 6px;
        height: 6px;
      }
      .file-table-wrapper .ant-table-body::-webkit-scrollbar-track {
        background: transparent;
      }
      .file-table-wrapper .ant-table-body::-webkit-scrollbar-thumb {
        background: transparent;
        border-radius: 3px;
      }
      .file-table-wrapper:hover .ant-table-body::-webkit-scrollbar-thumb {
        background: ${colors.border};
      }
      .file-table-wrapper .ant-table-body::-webkit-scrollbar-thumb:hover {
        background: ${colors.textTertiary};
      }
    `
    document.head.appendChild(style)
    return () => {
      const el = document.getElementById(id)
      if (el) el.remove()
    }
  }, [])

  return (
    <div ref={wrapperRef} className="file-table-wrapper" style={{ height: '100%' }}>
    <Table
      columns={columns}
      dataSource={dataSource as any[]}
      loading={loading}
      pagination={false}
      size="middle"
      showSorterTooltip={false}
      scroll={scrollY > 0 ? { y: scrollY } : undefined}
      locale={{ emptyText: '此文件夹为空' }}
      onRow={() => ({
        style: { cursor: 'default' },
      })}
    />
    </div>
  )
}
