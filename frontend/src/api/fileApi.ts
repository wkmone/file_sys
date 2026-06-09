import type { AxiosProgressEvent } from 'axios'
import apiClient from './client'
import type { ApiResponse, PaginatedResponse } from '../types/api'
import type { FileItem, FileVersion, BatchUploadResponse } from '../types/file'

export const fileApi = {
  list: (params: { folder_id?: string; team_id?: string; page?: number; page_size?: number }) =>
    apiClient.get<ApiResponse<PaginatedResponse<FileItem>>>('/files', { params }),

  upload: (formData: FormData) =>
    apiClient.post<ApiResponse<FileItem>>('/files', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  batchUpload: (formData: FormData, onProgress?: (e: AxiosProgressEvent) => void) =>
    apiClient.post<ApiResponse<BatchUploadResponse>>('/files/batch', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000,
      onUploadProgress: onProgress,
    }),

  get: (id: string) =>
    apiClient.get<ApiResponse<FileItem>>(`/files/${id}`),

  update: (id: string, data: { name?: string; folder_id?: string }) =>
    apiClient.patch<ApiResponse<null>>(`/files/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/files/${id}`),

  downloadUrl: (id: string) => {
    const token = localStorage.getItem('fs_access_token')
    const sep = token ? `?token=${encodeURIComponent(token)}` : ''
    return `/api/v1/files/${id}/download${sep}`
  },

  copy: (id: string, data: { folder_id: string }) =>
    apiClient.post<ApiResponse<null>>(`/files/${id}/copy`, data),

  share: (id: string, data: { user_id?: string; team_id?: string; permission: string }) =>
    apiClient.post<ApiResponse<null>>(`/files/${id}/share`, data),

  // Versions
  versions: (fileId: string) =>
    apiClient.get<ApiResponse<FileVersion[]>>(`/files/${fileId}/versions`),

  restoreVersion: (fileId: string, versionId: string) =>
    apiClient.post<ApiResponse<null>>(`/files/${fileId}/versions/${versionId}/restore`),
}
