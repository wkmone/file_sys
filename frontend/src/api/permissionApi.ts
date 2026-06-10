import apiClient from './client'
import type { ApiResponse } from '../types/api'

export const permissionApi = {
  update: (id: string, permission: string) =>
    apiClient.patch<ApiResponse<null>>(`/permissions/${id}`, { permission }),

  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/permissions/${id}`),
}
