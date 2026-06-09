export interface FolderBreadcrumbItem {
  id: string
  name: string
}

export interface Folder {
  id: string
  name: string
  parent_id: string | null
  owner_id: string
  team_id: string | null
  folder_path: string
  is_deleted: boolean
  breadcrumb?: FolderBreadcrumbItem[]
  created_at: string
  updated_at: string
}
