import { Row, Col, Empty } from 'antd'
import type { FileItem } from '../../types/file'
import FileCard from './FileCard'

interface FileListProps {
  files: FileItem[]
  onEdit: (file: FileItem) => void
  onRefresh: () => void
  onShare: (file: FileItem) => void
}

export default function FileList({ files, onEdit, onRefresh, onShare }: FileListProps) {
  if (files.length === 0) {
    return <Empty description="暂无文件" />
  }

  return (
    <Row gutter={[12, 12]}>
      {files.map((file) => (
        <Col key={file.id} span={24} sm={12} lg={8} xl={6}>
          <FileCard file={file} onEdit={onEdit} onDelete={onRefresh} onShare={onShare} />
        </Col>
      ))}
    </Row>
  )
}
