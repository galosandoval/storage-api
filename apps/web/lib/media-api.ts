import type {
  MediaItem,
  MediaListResponse,
  ErrorResponse,
  MediaTypeFilter
} from './types/media'

// Use local API routes to avoid CORS issues
// These routes proxy to the actual backend server-side

export async function listMedia(
  page = 1,
  pageSize = 20,
  typeFilter?: MediaTypeFilter
): Promise<MediaListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize)
  })

  if (typeFilter && typeFilter !== 'all') {
    params.set('type', typeFilter)
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

export function getThumbnailUrl(id: string): string {
  return `/api/media/${id}/thumbnail`
}

export async function getMediaBlob(id: string): Promise<string> {
  const response = await fetch(`/api/media/${id}/download`)

  if (!response.ok) {
    throw new Error('Failed to fetch media')
  }

  const blob = await response.blob()
  return URL.createObjectURL(blob)
}

export async function getThumbnailBlob(id: string): Promise<string> {
  const response = await fetch(`/api/media/${id}/thumbnail`)

  if (!response.ok) {
    // Fallback to full image if thumbnail not available
    return getMediaBlob(id)
  }

  const blob = await response.blob()
  return URL.createObjectURL(blob)
}

export interface UploadOptions {
  file: File
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
  onProgress
}: UploadOptions): Promise<UploadOutcome> {
  return new Promise((resolve) => {
    const xhr = new XMLHttpRequest()
    const formData = new FormData()
    formData.append('file', file)

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
