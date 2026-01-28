'use client'

import { useCallback, useRef, useState } from 'react'
import { Upload, X, AlertCircle, Lock, Globe } from 'lucide-react'
import { useMediaUpload } from '@/hooks/use-media-upload'
import type { MediaItem } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'

interface UploadDropzoneProps {
  onUploadComplete: (item: MediaItem) => void
}

export function UploadDropzone({ onUploadComplete }: UploadDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [isPrivate, setIsPrivate] = useState(false)
  const { isUploading, progress, error, upload, reset } = useMediaUpload()

  const handleFiles = useCallback(
    async (files: FileList | null) => {
      if (!files || files.length === 0) return

      // Process files one at a time
      for (const file of Array.from(files)) {
        const item = await upload({ file, isPrivate })
        if (item) {
          onUploadComplete(item)
        }
      }
    },
    [upload, onUploadComplete, isPrivate]
  )

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setIsDragging(false)
      handleFiles(e.dataTransfer.files)
    },
    [handleFiles]
  )

  const handleClick = useCallback(() => {
    inputRef.current?.click()
  }, [])

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      handleFiles(e.target.files)
      // Reset input so same file can be selected again
      e.target.value = ''
    },
    [handleFiles]
  )

  return (
    <div className='space-y-3'>
      {/* Privacy toggle */}
      <div className='flex items-center justify-between px-1'>
        <div className='flex items-center gap-2 text-sm'>
          {isPrivate ? (
            <>
              <Lock className='size-4 text-muted-foreground' />
              <span className='text-muted-foreground'>Private</span>
            </>
          ) : (
            <>
              <Globe className='size-4 text-muted-foreground' />
              <span className='text-muted-foreground'>Visible to household</span>
            </>
          )}
        </div>
        <div className='flex items-center gap-2'>
          <span className='text-xs text-muted-foreground'>Private</span>
          <Switch
            checked={isPrivate}
            onCheckedChange={setIsPrivate}
            disabled={isUploading}
          />
        </div>
      </div>

      <div
        role='button'
        tabIndex={0}
        onClick={handleClick}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            handleClick()
          }
        }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={cn(
          'relative flex flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 transition-colors cursor-pointer',
          'hover:border-primary/50 hover:bg-accent/50',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          isDragging && 'border-primary bg-accent',
          isUploading && 'pointer-events-none opacity-60'
        )}
      >
        <input
          ref={inputRef}
          type='file'
          accept='image/*,video/*'
          multiple
          onChange={handleChange}
          className='sr-only'
          disabled={isUploading}
        />

        <Upload className='size-10 text-muted-foreground' />
        <div className='text-center'>
          <p className='font-medium'>
            {isDragging ? 'Drop files here' : 'Drop files or click to upload'}
          </p>
          <p className='text-sm text-muted-foreground mt-1'>
            Photos and videos up to 100MB
          </p>
        </div>
      </div>

      {/* Upload progress */}
      {isUploading && (
        <div className='flex items-center gap-3'>
          <Progress value={progress} className='flex-1' />
          <span className='text-sm text-muted-foreground tabular-nums w-12'>
            {progress}%
          </span>
        </div>
      )}

      {/* Error message */}
      {error && (
        <div className='flex items-center justify-between gap-2 rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive'>
          <div className='flex items-center gap-2'>
            <AlertCircle className='size-4 shrink-0' />
            <span>{error}</span>
          </div>
          <Button
            variant='ghost'
            size='icon'
            className='size-6 text-destructive hover:text-destructive'
            onClick={reset}
          >
            <X className='size-4' />
          </Button>
        </div>
      )}
    </div>
  )
}
