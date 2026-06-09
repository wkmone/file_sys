export interface FileItem {
  id: string
  name: string
  original_name: string
  folder_id: string | null
  owner_id: string
  owner_name: string
  mime_type: string
  file_size: number
  file_ext: string
  current_version: number
  is_deleted: boolean
  created_at: string
  updated_at: string
}

export interface FileVersion {
  id: string
  file_id: string
  version_number: number
  file_size: number
  content_hash: string
  created_by: string | null
  change_note: string | null
  created_at: string
}

export interface BatchUploadManifestEntry {
  path: string
  size: number
  is_directory: boolean
}

export interface BatchUploadFileResult {
  path: string
  file_id?: string
  status: 'success' | 'failed'
  error?: string
}

export interface BatchUploadResponse {
  total: number
  succeeded: number
  failed: number
  results: BatchUploadFileResult[]
}
