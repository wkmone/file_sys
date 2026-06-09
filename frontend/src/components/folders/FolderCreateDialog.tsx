import { useState } from 'react'
import { Modal, Input, message } from 'antd'
import { folderApi } from '../../api/folderApi'

interface FolderCreateDialogProps {
  open: boolean
  parentId: string | null
  teamId?: string | null
  onClose: () => void
  onSuccess: () => void
}

export default function FolderCreateDialog({ open, parentId, teamId, onClose, onSuccess }: FolderCreateDialogProps) {
  const [name, setName] = useState('')
  const [loading, setLoading] = useState(false)

  const handleOk = async () => {
    if (!name.trim()) return
    setLoading(true)
    try {
      await folderApi.create({ name: name.trim(), parent_id: parentId || undefined, team_id: teamId || undefined })
      message.success('文件夹已创建')
      setName('')
      onSuccess()
      onClose()
    } catch {
      message.error('创建失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="新建文件夹"
      open={open}
      onCancel={onClose}
      onOk={handleOk}
      confirmLoading={loading}
      okText="创建"
    >
      <Input
        placeholder="文件夹名称"
        value={name}
        onChange={(e) => setName(e.target.value)}
        onPressEnter={handleOk}
      />
    </Modal>
  )
}
