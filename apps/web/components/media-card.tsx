'use client'

import { useState, useEffect } from 'react'
import { MoreVertical, Trash2, Play } from 'lucide-react'
import { getMediaBlob } from '@/lib/media-api'
import type { MediaItem } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

interface MediaCardProps {
  item: MediaItem
  onClick: () => void
  onDelete: () => void
}

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

export function MediaCard({ item, onClick, onDelete }: MediaCardProps) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function loadMedia() {
      try {
        setIsLoading(true)
        setError(false)
        const url = await getMediaBlob(item.id)
        if (!cancelled) {
          setBlobUrl(url)
        }
      } catch {
        if (!cancelled) {
          setError(true)
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    loadMedia()

    return () => {
      cancelled = true
      if (blobUrl) {
        URL.revokeObjectURL(blobUrl)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [item.id])

  return (
    <div className='group relative aspect-square overflow-hidden rounded-lg bg-muted'>
      {/* Media content */}
      <button
        type='button'
        onClick={onClick}
        className='absolute inset-0 w-full h-full cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2'
      >
        {isLoading && (
          <div className='absolute inset-0 flex items-center justify-center'>
            <div className='size-8 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent' />
          </div>
        )}

        {error && (
          <div className='absolute inset-0 flex items-center justify-center text-muted-foreground text-sm'>
            Failed to load
          </div>
        )}

        {blobUrl && !error && (
          <>
            {item.type === 'photo' ? (
              <img
                src={blobUrl}
                alt=''
                className={cn(
                  'absolute inset-0 w-full h-full object-cover transition-transform duration-200',
                  'group-hover:scale-105'
                )}
                loading='lazy'
              />
            ) : (
              <video
                src={blobUrl}
                className={cn(
                  'absolute inset-0 w-full h-full object-cover transition-transform duration-200',
                  'group-hover:scale-105'
                )}
                muted
                playsInline
                preload='metadata'
              />
            )}
          </>
        )}

        {/* Video indicator */}
        {item.type === 'video' && !isLoading && !error && (
          <>
            <div className='absolute inset-0 flex items-center justify-center'>
              <div className='rounded-full bg-black/50 p-3'>
                <Play className='size-6 text-white fill-white' />
              </div>
            </div>
            {item.durationSec && (
              <div className='absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white'>
                {formatDuration(item.durationSec)}
              </div>
            )}
          </>
        )}
      </button>

      {/* Action menu */}
      <div className='absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity'>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='secondary'
              size='icon'
              className='size-8 bg-black/50 hover:bg-black/70 text-white'
              onClick={(e) => e.stopPropagation()}
            >
              <MoreVertical className='size-4' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            <DropdownMenuItem
              className='text-destructive focus:text-destructive'
              onClick={(e) => {
                e.stopPropagation()
                onDelete()
              }}
            >
              <Trash2 className='size-4 mr-2' />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
