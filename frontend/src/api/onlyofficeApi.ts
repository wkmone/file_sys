import apiClient from './client'
import type { ApiResponse } from '../types/api'

export interface EditorConfig {
  token: string
  config?: Record<string, any>
}

export const onlyofficeApi = {
  getEditorConfig: (fileId: string, mode: 'edit' | 'view') =>
    apiClient.post<ApiResponse<EditorConfig>>('/oo/editor-config', { file_id: fileId, mode }),

  getDocServerUrl: () => import.meta.env.VITE_ONLYOFFICE_DS_URL || 'http://localhost:9980',
}
