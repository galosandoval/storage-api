import type {
  MediaItem,
  MediaListResponse,
  ErrorResponse,
  MediaTypeFilter,
  VisibilityFilter
} from './types/media'

// Use local API routes to avoid CORS issues
// These routes proxy to the actual backend server-side

export async function listMedia(
  page = 1,
  pageSize = 20,
  typeFilter?: MediaTypeFilter,
  visibilityFilter?: VisibilityFilter
): Promise<MediaListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize)
  })

  if (typeFilter && typeFilter !== 'all') {
    params.set('type', typeFilter)
  }

  if (visibilityFilter) {
    params.set('visibility', visibilityFilter)
  }

  const response = await fetch(`/api/media?${params}`)

  if (!response.ok) {
    const error: ErrorResponse = await response.json()
    throw new Error(error.error)
  }

  return response.json()
}

export async function getMedia(id: string): Promise<MediaItem> {
  const response = await fetch(`/api/media/${id}`)

  if (!response.ok) {
    const error: ErrorResponse = await response.json()
    throw new Error(error.error)
  }

  const data = await response.json()
  return data.item
}

export function getMediaUrl(id: string): string {
  return `/api/media/${id}/download`
}

/**
 *
 * @param id - The ID of the media item
 * @returns The URL of the original media file
 */
export function getOriginalMediaUrl(id: string): string {
  return `/api/media/${id}/original`
}

export function getThumbnailUrl(id: string): string {
  return `/api/media/${id}/thumbnail`
}

export interface UploadOptions {
  file: File
  isPrivate?: boolean
  onProgress?: (progress: number) => void
}

export interface UploadResult {
  success: true
  item: MediaItem
}

export interface UploadConflict {
  success: false
  conflict: true
  existing: MediaItem
}

export interface UploadError {
  success: false
  conflict: false
  error: string
}

export type UploadOutcome = UploadResult | UploadConflict | UploadError

export function uploadMedia({
  file,
  isPrivate = false,
  onProgress
}: UploadOptions): Promise<UploadOutcome> {
  return new Promise((resolve) => {
    const xhr = new XMLHttpRequest()
    const formData = new FormData()
    formData.append('file', file)
    if (isPrivate) {
      formData.append('is_private', 'true')
    }

    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable && onProgress) {
        const progress = Math.round((event.loaded / event.total) * 100)
        onProgress(progress)
      }
    })

    xhr.addEventListener('load', async () => {
      try {
        const data = JSON.parse(xhr.responseText)

        if (xhr.status === 201) {
          resolve({ success: true, item: data.item })
        } else if (xhr.status === 409) {
          resolve({ success: false, conflict: true, existing: data.existing })
        } else {
          resolve({
            success: false,
            conflict: false,
            error: data.error || 'Upload failed'
          })
        }
      } catch {
        resolve({
          success: false,
          conflict: false,
          error: 'Failed to parse response'
        })
      }
    })

    xhr.addEventListener('error', () => {
      resolve({ success: false, conflict: false, error: 'Network error' })
    })

    // Use local API route to avoid CORS
    xhr.open('POST', '/api/media')
    xhr.send(formData)
  })
}

export async function deleteMedia(id: string): Promise<void> {
  const response = await fetch(`/api/media/${id}`, {
    method: 'DELETE'
  })

  if (!response.ok) {
    const error: ErrorResponse = await response.json()
    throw new Error(error.error)
  }
}
