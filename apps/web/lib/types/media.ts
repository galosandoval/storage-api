export interface MediaItem {
  id: string
  householdId: string
  uploaderId?: string
  isPrivate: boolean
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

  // Preview, thumbnail, and original file paths
  previewPath?: string
  thumbnailPath?: string
  originalFilename?: string

  // Camera metadata (from EXIF)
  cameraMake?: string
  cameraModel?: string

  // GPS coordinates (from EXIF)
  latitude?: number
  longitude?: number

  // Technical metadata (from EXIF)
  orientation?: number
  iso?: number
  fNumber?: number
  exposureTime?: string
  focalLength?: number
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

export type VisibilityFilter = 'all' | 'mine'
