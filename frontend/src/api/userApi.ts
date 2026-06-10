import apiClient from './client'
import type { ApiResponse } from '../types/api'
import type { User } from '../types/user'

export const userApi = {
  search: (query: string) =>
    apiClient.get<ApiResponse<User[]>>('/users/search', { params: { q: query } }),
}
