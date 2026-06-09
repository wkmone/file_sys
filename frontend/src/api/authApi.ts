import apiClient from './client'
import type { ApiResponse } from '../types/api'
import type { AuthResponse, User } from '../types/user'

interface PaginatedUsers {
  items: User[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const authApi = {
  register: (data: { email: string; password: string; display_name: string }) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data),

  login: (data: { email: string; password: string }) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data),

  refresh: () =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/refresh'),

  logout: () =>
    apiClient.post<ApiResponse<null>>('/auth/logout'),

  me: () =>
    apiClient.get<ApiResponse<User>>('/auth/me'),

  changePassword: (data: { old_password: string; new_password: string }) =>
    apiClient.patch<ApiResponse<null>>('/auth/password', data),

  listUsers: (params?: { page?: number; page_size?: number }) =>
    apiClient.get<ApiResponse<PaginatedUsers>>('/users', { params }),
}
