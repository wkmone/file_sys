import apiClient from './client'
import type { ApiResponse, PaginatedResponse } from '../types/api'
import type { Folder } from '../types/folder'

export const folderApi = {
  list: (params: { parent_id?: string; team_id?: string; page?: number; page_size?: number }) =>
    apiClient.get<ApiResponse<PaginatedResponse<Folder>>>('/folders', { params }),

  create: (data: { name: string; parent_id?: string; team_id?: string }) =>
    apiClient.post<ApiResponse<Folder>>('/folders', data),

  get: (id: string) =>
    apiClient.get<ApiResponse<Folder>>(`/folders/${id}`),

  update: (id: string, data: { name?: string; parent_id?: string }) =>
    apiClient.patch<ApiResponse<null>>(`/folders/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/folders/${id}`),

  tree: () =>
    apiClient.get<ApiResponse<Folder[]>>('/folders?tree=true'),

  share: (id: string, data: { user_id?: string; team_id?: string; permission: string }) =>
    apiClient.post<ApiResponse<null>>(`/folders/${id}/share`, data),
}

export const folderPermissionApi = {
  list: (folderId: string) =>
    apiClient.get<ApiResponse<import('../types/file').FilePermission[]>>(`/folders/${folderId}/permissions`),

  sharedWithMe: () =>
    apiClient.get<ApiResponse<import('../types/folder').Folder[]>>('/folders/shared-with-me'),
}
