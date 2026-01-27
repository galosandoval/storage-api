'use client'

import { useCallback, useState } from 'react'
import { uploadMedia, type UploadOutcome } from '@/lib/media-api'
import type { MediaItem } from '@/lib/types/media'

const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100MB

interface UploadState {
  isUploading: boolean
  progress: number
  error: string | null
  conflict: MediaItem | null
}

interface UploadFileOptions {
  file: File
  isPrivate?: boolean
}

interface UseMediaUploadReturn extends UploadState {
  upload: (options: UploadFileOptions) => Promise<MediaItem | null>
  reset: () => void
}

export function useMediaUpload(): UseMediaUploadReturn {
  const [state, setState] = useState<UploadState>({
    isUploading: false,
    progress: 0,
    error: null,
    conflict: null
  })

  const reset = useCallback(() => {
    setState({
      isUploading: false,
      progress: 0,
      error: null,
      conflict: null
    })
  }, [])

  const upload = useCallback(
    async ({ file, isPrivate }: UploadFileOptions): Promise<MediaItem | null> => {
      // Validate file size
      if (file.size > MAX_FILE_SIZE) {
        setState((prev) => ({
          ...prev,
          error: `File too large. Maximum size is 100MB.`
        }))
        return null
      }

      // Validate file type
      const isValidType =
        file.type.startsWith('image/') || file.type.startsWith('video/')
      if (!isValidType) {
        setState((prev) => ({
          ...prev,
          error: 'Only images and videos are allowed.'
        }))
        return null
      }

      setState({
        isUploading: true,
        progress: 0,
        error: null,
        conflict: null
      })

      const result: UploadOutcome = await uploadMedia({
        file,
        isPrivate,
        onProgress: (progress) => {
          setState((prev) => ({ ...prev, progress }))
        }
      })

      if (result.success) {
        setState({
          isUploading: false,
          progress: 100,
          error: null,
          conflict: null
        })
        return result.item
      }

      if (result.conflict) {
        setState({
          isUploading: false,
          progress: 0,
          error: 'This file already exists.',
          conflict: result.existing
        })
        return null
      }

      setState({
        isUploading: false,
        progress: 0,
        error: result.error,
        conflict: null
      })
      return null
    },
    []
  )

  return {
    ...state,
    upload,
    reset
  }
}
