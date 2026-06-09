export interface User {
  id: string
  email: string
  display_name: string
  avatar_url: string | null
  role: 'super_admin' | 'admin' | 'member'
  created_at: string
}

export interface AuthResponse {
  access_token: string
  expires_in: number
  user: User
}
