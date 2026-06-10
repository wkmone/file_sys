import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { Button, Space, Breadcrumb, Spin, Segmented, Input } from 'antd'
import {
  UploadOutlined,
  FolderAddOutlined,
  FileAddOutlined,
  HomeOutlined,
  UnorderedListOutlined,
  AppstoreOutlined,
  SearchOutlined,
  TeamOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import FileList from '../components/files/FileList'
import FileTable from '../components/files/FileTable'
import FileUpload from '../components/files/FileUpload'
import FolderCreateDialog from '../components/folders/FolderCreateDialog'
import FileCreateDialog from '../components/files/FileCreateDialog'
import ShareDialog from '../components/files/ShareDialog'
import { fileApi } from '../api/fileApi'
import { folderApi } from '../api/folderApi'
import { teamApi } from '../api/teamApi'
import type { FileItem } from '../types/file'
import type { Folder } from '../types/folder'
import { colors, OFFICE_EDITABLE_EXTS } from '../theme'

export default function FileBrowserPage() {
  const { folderId, teamId } = useParams<{ folderId?: string; teamId?: string }>()
  const navigate = useNavigate()
  const location = useLocation()

  // Determine scope from URL path
  const scope = useMemo<'personal' | 'team'>(() => {
    return location.pathname.startsWith('/teams/') ? 'team' : 'personal'
  }, [location.pathname])

  const [files, setFiles] = useState<FileItem[]>([])
  const [folders, setFolders] = useState<Folder[]>([])
  const [loading, setLoading] = useState(false)
  const [uploadOpen, setUploadOpen] = useState(false)
  const [createFolderOpen, setCreateFolderOpen] = useState(false)
  const [createFileOpen, setCreateFileOpen] = useState(false)
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table')
  const [breadcrumbPath, setBreadcrumbPath] = useState<{ id: string; name: string }[]>([])
  const internalNav = useRef(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [teamName, setTeamName] = useState('')
  const [shareOpen, setShareOpen] = useState(false)
  const [shareFile, setShareFile] = useState<FileItem | null>(null)

  // Load team name for breadcrumb
  useEffect(() => {
    if (scope === 'team' && teamId) {
      teamApi.get(teamId).then((res) => setTeamName(res.data.data?.name || ''))
        .catch(() => {})
    }
  }, [scope, teamId])

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const params: any = { page_size: 10000 }
      if (scope === 'personal') {
        params.folder_id = folderId || undefined
        params.parent_id = folderId || undefined
      } else if (scope === 'team' && teamId) {
        params.folder_id = folderId || undefined
        params.parent_id = folderId || undefined
        params.team_id = teamId
      }

      const [fileRes, folderRes] = await Promise.all([
        fileApi.list(params),
        folderApi.list(params),
      ])
      setFiles(fileRes.data.data.items || [])
      setFolders(folderRes.data.data.items || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [folderId, scope, teamId])

  useEffect(() => { fetchData() }, [fetchData])

  // Breadcrumb reset on root; direct URL access fallback
  useEffect(() => {
    if (!folderId) {
      setBreadcrumbPath([])
      return
    }
    if (internalNav.current) {
      internalNav.current = false
      return
    }
    // Direct URL access: fetch folder name and reset breadcrumb
    folderApi.get(folderId).then((res) => {
      const folder = res.data.data
      setBreadcrumbPath(folder.breadcrumb && folder.breadcrumb.length > 1
        ? folder.breadcrumb
        : [{ id: folder.id, name: folder.name }])
    }).catch(() => {})
  }, [folderId])

  const filteredFiles = searchQuery
    ? files.filter((f) => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
    : files
  const filteredFolders = searchQuery
    ? folders.filter((f) => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
    : folders

  const handleEnterFolder = (folder: Folder) => {
    internalNav.current = true
    setBreadcrumbPath((prev) => [...prev, { id: folder.id, name: folder.name }])
    if (scope === 'team' && teamId) {
      navigate(`/teams/${teamId}/files/${folder.id}`)
    } else {
      navigate(`/my/files/${folder.id}`)
    }
  }

  const handleEditFile = (file: FileItem) => {
    const isOffice = OFFICE_EDITABLE_EXTS.includes(file.file_ext)
    if (isOffice) {
      window.open(`/editor/${file.id}`, '_blank')
    } else {
      window.open(fileApi.downloadUrl(file.id), '_blank')
    }
  }

  const handleShare = (file: FileItem) => {
    setShareFile(file)
    setShareOpen(true)
  }

  // Build breadcrumb items
  const breadcrumbItems = (() => {
    const items: any[] = [
      {
        title: <HomeOutlined style={{ cursor: 'pointer', color: colors.primary }} />,
        onClick: () => navigate('/'),
      },
    ]

    if (scope === 'personal') {
      items.push({
        title: <span style={{ cursor: 'pointer', color: colors.textSecondary }} onClick={() => navigate('/my/files')}>个人空间</span>,
      })
    } else if (scope === 'team' && teamId) {
      items.push({
        title: (
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <TeamOutlined />
            <span style={{ cursor: 'pointer', color: colors.textSecondary }} onClick={() => navigate(`/teams/${teamId}/files`)}>
              {teamName || '团队空间'}
            </span>
          </span>
        ),
      })
    }

    breadcrumbPath.forEach((item, i) => {
      const isLast = i === breadcrumbPath.length - 1
      const basePath = scope === 'team' && teamId ? `/teams/${teamId}/files` : '/my/files'
      items.push({
        title: isLast
          ? <span style={{ fontWeight: 500, color: colors.textPrimary }}>{item.name}</span>
          : <span
              style={{ cursor: 'pointer', color: colors.textSecondary }}
              onClick={() => {
                internalNav.current = true
                setBreadcrumbPath((prev) => prev.slice(0, i + 1))
                navigate(`${basePath}/${item.id}`)
              }}
            >{item.name}</span>,
      })
    })

    return items
  })()

  const toolbarButtons = (
    <>
      <Button size="small" icon={<FolderAddOutlined />} onClick={() => setCreateFolderOpen(true)}>
        新建文件夹
      </Button>
      <Button size="small" icon={<FileAddOutlined />} onClick={() => setCreateFileOpen(true)}>
        新建文件
      </Button>
      <Button size="small" type="primary" icon={<UploadOutlined />} onClick={() => setUploadOpen(true)}>
        上传文件
      </Button>
    </>
  )

  return (
    <div style={{ margin: -24, display: 'flex', flexDirection: 'column', height: 'calc(100vh - 192px)' }}>
      {/* Toolbar */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '12px 24px',
        borderBottom: `1px solid ${colors.border}`,
        background: colors.white,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <Breadcrumb items={breadcrumbItems} style={{ fontSize: 14 }} />
        </div>
        <Space size="middle">
          <Input
            prefix={<SearchOutlined style={{ color: colors.textTertiary }} />}
            placeholder="在当前目录搜索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            allowClear
            style={{ width: 220 }}
            size="small"
          />
          <Segmented
            size="small"
            value={viewMode}
            onChange={(val) => setViewMode(val as 'table' | 'grid')}
            options={[
              { label: '', value: 'table', icon: <UnorderedListOutlined /> },
              { label: '', value: 'grid', icon: <AppstoreOutlined /> },
            ]}
          />
          <Button size="small" icon={<ReloadOutlined />} onClick={fetchData} loading={loading} />
          {toolbarButtons}
        </Space>
      </div>

      {/* File/Folder List */}
      {viewMode === 'table' ? (
        <div style={{ flex: 1, padding: 20, overflow: 'hidden' }}>
          <FileTable
            files={filteredFiles}
            folders={filteredFolders}
            loading={loading}
            onEnterFolder={handleEnterFolder}
            onEditFile={handleEditFile}
            onRefresh={fetchData}
            onShare={handleShare}
          />
        </div>
      ) : (
        <div style={{ flex: 1, overflow: 'auto', padding: 20 }}>
          <>
            {filteredFolders.length > 0 && (
              <div style={{ marginBottom: 20 }}>
                <div style={{ fontSize: 12, color: colors.textTertiary, marginBottom: 8, fontWeight: 500 }}>
                  文件夹
                </div>
                <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
                  {filteredFolders.map((f) => (
                    <div
                      key={f.id}
                      onClick={() => handleEnterFolder(f)}
                      style={{
                        padding: '10px 14px',
                        border: `1px solid ${colors.border}`,
                        borderRadius: 8,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        fontSize: 14,
                        background: colors.white,
                        transition: 'all 0.15s',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = colors.primary
                        e.currentTarget.style.background = colors.primaryLight
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = colors.border
                        e.currentTarget.style.background = colors.white
                      }}
                    >
                      <FolderAddOutlined style={{ color: '#f5a623' }} />
                      <span>{f.name}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <Spin spinning={loading}>
              <FileList files={filteredFiles} onEdit={handleEditFile} onRefresh={fetchData} onShare={handleShare} />
            </Spin>
          </>
        </div>
      )}

      <FileUpload
        open={uploadOpen}
        folderId={folderId || null}
        teamId={scope === 'team' ? teamId : null}
        onClose={() => setUploadOpen(false)}
        onSuccess={fetchData}
      />

      <FolderCreateDialog
        open={createFolderOpen}
        parentId={folderId || null}
        teamId={scope === 'team' ? teamId : null}
        onClose={() => setCreateFolderOpen(false)}
        onSuccess={fetchData}
      />

      <FileCreateDialog
        open={createFileOpen}
        folderId={folderId || null}
        teamId={scope === 'team' ? teamId : null}
        onClose={() => setCreateFileOpen(false)}
        onSuccess={fetchData}
      />

      <ShareDialog
        open={shareOpen}
        resourceType="file"
        resourceId={shareFile?.id || ''}
        resourceName={shareFile?.name || ''}
        scope={scope}
        onClose={() => { setShareOpen(false); setShareFile(null) }}
      />
    </div>
  )
}
