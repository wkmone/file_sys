import { useState } from 'react'
import { Input, Modal, List, Typography, Tag } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import type { FileItem } from '../../types/file'
import type { Folder } from '../../types/folder'
import apiClient from '../../api/client'
import FileIcon from '../common/FileIcon'

export default function SearchBar() {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(false)
  const [files, setFiles] = useState<FileItem[]>([])
  const [folders, setFolders] = useState<Folder[]>([])

  const handleSearch = async (q: string) => {
    setQuery(q)
    if (q.length < 1) {
      setFiles([])
      setFolders([])
      return
    }
    setLoading(true)
    try {
      const res = await apiClient.get('/search', { params: { q } })
      setFiles(res.data.data.files || [])
      setFolders(res.data.data.folders || [])
    } catch {
      /* ignore */
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Input
        prefix={<SearchOutlined />}
        placeholder="搜索文件..."
        style={{ width: 280 }}
        onFocus={() => setOpen(true)}
        value={query}
        onChange={(e) => handleSearch(e.target.value)}
      />
      <Modal
        title="搜索结果"
        open={open && query.length > 0}
        onCancel={() => { setOpen(false); setQuery('') }}
        footer={null}
        width={600}
      >
        <List
          loading={loading}
          locale={{ emptyText: query ? '未找到匹配的文件或文件夹' : '输入关键词搜索' }}
        >
          {folders.map((f) => (
            <List.Item key={`folder-${f.id}`}>
              <List.Item.Meta
                avatar={<FileIcon type="folder" size={28} />}
                title={f.name}
              />
              <Tag>文件夹</Tag>
            </List.Item>
          ))}
          {files.map((f) => (
            <List.Item key={f.id}>
              <List.Item.Meta
                avatar={<FileIcon type={f.file_ext || ''} size={28} />}
                title={f.name}
                description={f.file_ext}
              />
              <Tag color="blue">
                {f.file_size > 1024 * 1024
                  ? (f.file_size / (1024 * 1024)).toFixed(1) + ' MB'
                  : (f.file_size / 1024).toFixed(1) + ' KB'}
              </Tag>
            </List.Item>
          ))}
        </List>
      </Modal>
    </>
  )
}
