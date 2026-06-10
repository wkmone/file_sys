import { useState } from 'react'
import { Modal, Input, Segmented, message } from 'antd'
import { fileApi } from '../../api/fileApi'

interface FileCreateDialogProps {
  open: boolean
  folderId: string | null
  teamId?: string | null
  onClose: () => void
  onSuccess: () => void
}

const FILE_TYPES: { label: string; value: string }[] = [
  { label: 'Word 文档', value: '.docx' },
  { label: 'Excel 表格', value: '.xlsx' },
  { label: 'PPT 演示', value: '.pptx' },
]

export default function FileCreateDialog({ open, folderId, teamId, onClose, onSuccess }: FileCreateDialogProps) {
  const [name, setName] = useState('')
  const [fileExt, setFileExt] = useState('.docx')
  const [loading, setLoading] = useState(false)

  const handleOk = async () => {
    if (!name.trim()) return
    setLoading(true)
    try {
      await fileApi.createBlank(
        { name: name.trim(), file_ext: fileExt, folder_id: folderId || undefined },
        teamId || undefined,
      )
      message.success('文件已创建')
      setName('')
      onSuccess()
      onClose()
    } catch {
      message.error('创建失败')
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setName('')
    setFileExt('.docx')
    onClose()
  }

  return (
    <Modal
      title="新建文件"
      open={open}
      onCancel={handleClose}
      onOk={handleOk}
      confirmLoading={loading}
      okText="创建"
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Segmented
          block
          options={FILE_TYPES}
          value={fileExt}
          onChange={(v) => setFileExt(v as string)}
        />
        <Input
          placeholder="文件名称（不含扩展名）"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onPressEnter={handleOk}
        />
      </div>
    </Modal>
  )
}
