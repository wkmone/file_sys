import { useParams } from 'react-router-dom'
import { Typography } from 'antd'
import OnlyOfficeEditor from '../components/editor/OnlyOfficeEditor'

export default function EditorPage() {
  const { fileId } = useParams<{ fileId: string }>()

  if (!fileId) {
    return <Typography.Text type="danger">文件 ID 缺失</Typography.Text>
  }

  return (
    <div style={{ width: '100vw', height: '100vh' }}>
      <OnlyOfficeEditor fileId={fileId} mode="edit" />
    </div>
  )
}
