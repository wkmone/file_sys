import { useState, useEffect, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Tree, Spin } from 'antd'
import { HomeOutlined } from '@ant-design/icons'
import type { Folder } from '../../types/folder'
import { folderApi } from '../../api/folderApi'
import FileIcon from '../common/FileIcon'
import { colors } from '../../theme'

interface FolderTreeProps {
  currentFolderId: string | null
}

interface TreeNode {
  title: string
  key: string
  icon: React.ReactNode
  children: TreeNode[]
  isLeaf: boolean
}

function buildTree(folders: Folder[], parentId: string | null): TreeNode[] {
  return folders
    .filter((f) => f.parent_id === parentId)
    .map((f) => ({
      title: f.name,
      key: f.id,
      icon: <FileIcon type="folder" size={22} />,
      children: buildTree(folders, f.id),
      isLeaf: false,
    }))
}

export default function FolderTree({ currentFolderId }: FolderTreeProps) {
  const navigate = useNavigate()
  const [folders, setFolders] = useState<Folder[]>([])
  const [loading, setLoading] = useState(false)
  const [expandedKeys, setExpandedKeys] = useState<string[]>([])

  useEffect(() => {
    setLoading(true)
    folderApi.tree()
      .then((res) => {
        const data = res.data?.data
        setFolders(Array.isArray(data) ? data : [])
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  // Expand folders to reveal current selection
  useEffect(() => {
    if (currentFolderId && folders.length > 0) {
      const parentIds: string[] = []
      let current = folders.find((f) => f.id === currentFolderId)
      while (current?.parent_id) {
        parentIds.push(current.parent_id)
        current = folders.find((f) => f.id === current!.parent_id)
      }
      setExpandedKeys((prev) => {
        const merged = new Set([...prev, ...parentIds])
        return Array.from(merged)
      })
    }
  }, [currentFolderId, folders])

  const treeData: TreeNode[] = useMemo(() => {
    const children = buildTree(folders, null)
    return [
      {
        title: '全部文件',
        key: 'root',
        icon: <HomeOutlined style={{ color: colors.primary }} />,
        children,
        isLeaf: false,
      },
    ]
  }, [folders])

  if (loading) return <div style={{ padding: 24, textAlign: 'center' }}><Spin size="small" /></div>

  return (
    <Tree
      showIcon
      treeData={treeData}
      selectedKeys={currentFolderId ? [currentFolderId] : ['root']}
      expandedKeys={expandedKeys}
      onExpand={(keys) => setExpandedKeys(keys as string[])}
      onSelect={(keys) => {
        if (keys.length === 0) return
        const key = keys[0] as string
        navigate(key === 'root' ? '/files' : `/files/${key}`)
      }}
      style={{ background: 'transparent', fontSize: 14 }}
    />
  )
}
