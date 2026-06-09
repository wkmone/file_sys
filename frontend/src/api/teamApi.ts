import apiClient from './client'
import type { ApiResponse } from '../types/api'
import type { Team, TeamMember, JoinRequest } from '../types/team'

export const teamApi = {
  list: () =>
    apiClient.get<ApiResponse<Team[]>>('/teams'),

  discover: () =>
    apiClient.get<ApiResponse<Team[]>>('/teams/discover'),

  create: (data: { name: string; description: string }) =>
    apiClient.post<ApiResponse<Team>>('/teams', data),

  get: (id: string) =>
    apiClient.get<ApiResponse<Team>>(`/teams/${id}`),

  update: (id: string, data: { name?: string; description?: string }) =>
    apiClient.patch<ApiResponse<null>>(`/teams/${id}`, data),

  delete: (id: string) =>
    apiClient.delete<ApiResponse<null>>(`/teams/${id}`),

  members: (teamId: string) =>
    apiClient.get<ApiResponse<TeamMember[]>>(`/teams/${teamId}/members`),

  addMember: (teamId: string, data: { user_id: string; role: string }) =>
    apiClient.post<ApiResponse<null>>(`/teams/${teamId}/members`, data),

  updateMember: (teamId: string, userId: string, data: { role: string }) =>
    apiClient.patch<ApiResponse<null>>(`/teams/${teamId}/members/${userId}`, data),

  removeMember: (teamId: string, userId: string) =>
    apiClient.delete<ApiResponse<null>>(`/teams/${teamId}/members/${userId}`),

  requestJoin: (teamId: string) =>
    apiClient.post<ApiResponse<null>>(`/teams/${teamId}/join`),

  getPendingRequest: (teamId: string) =>
    apiClient.get<ApiResponse<JoinRequest>>(`/teams/${teamId}/pending`),

  listJoinRequests: (teamId: string) =>
    apiClient.get<ApiResponse<JoinRequest[]>>(`/teams/${teamId}/requests`),

  handleJoinRequest: (teamId: string, requestId: string, status: 'approved' | 'rejected') =>
    apiClient.patch<ApiResponse<null>>(`/teams/${teamId}/requests/${requestId}`, { status }),
}
