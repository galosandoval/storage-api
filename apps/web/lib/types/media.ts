export interface MediaItem {
  id: string
  householdId: string
  path: string
  type: 'photo' | 'video'
  mimeType?: string
  sizeBytes?: number
  sha256?: string
  takenAt?: string
  width?: number
  height?: number
  durationSec?: number
  createdAt: string
  updatedAt: string
}

export interface MediaListResponse {
  items: MediaItem[]
  totalCount: number
  page: number
  pageSize: number
}

export interface UploadResponse {
  message: string
  item: MediaItem
}

export interface ErrorResponse {
  error: string
  existing?: MediaItem // Only on 409 conflict
}

export type MediaTypeFilter = 'all' | 'photo' | 'video'
