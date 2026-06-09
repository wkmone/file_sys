import { create } from 'zustand'

// Lightweight event bus — components increment version after team mutations,
// and subscribers (sidebar, team pages) refetch when they observe a change.
interface TeamRefreshState {
  version: number
  bump: () => void
}

export const useTeamRefresh = create<TeamRefreshState>((set) => ({
  version: 0,
  bump: () => set((s) => ({ version: s.version + 1 })),
}))
