export interface Team {
  id: string
  name: string
  description: string | null
  owner_id: string
  created_at: string
  updated_at: string
}

export interface TeamMember {
  id: string
  team_id: string
  user_id: string
  display_name: string
  email: string
  role: 'owner' | 'admin' | 'member'
  joined_at: string
}

export interface JoinRequest {
  id: string
  team_id: string
  user_id: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
  updated_at: string
  display_name?: string
  email?: string
  team_name?: string
}
