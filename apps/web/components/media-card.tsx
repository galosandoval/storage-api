'use client'

import { useState } from 'react'
import Image from 'next/image'
import Link from 'next/link'
import { MoreVertical, Trash2, Play, Camera, Calendar } from 'lucide-react'
import type { MediaItem } from '@/lib/types/media'
import { getThumbnailUrl } from '@/lib/media-api'
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
  onDelete: () => void
}

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function formatDate(dateString?: string): string | null {
  if (!dateString) return null
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric'
    })
  } catch {
    return null
  }
}

export function MediaCard({ item, onDelete }: MediaCardProps) {
  const [isLoading, setIsLoading] = useState(true)
  const [isError, setIsError] = useState(false)

  const thumbnailUrl = getThumbnailUrl(item.id)
  const displayDate = formatDate(item.takenAt) || formatDate(item.createdAt)
  const cameraInfo = item.cameraModel || item.cameraMake

  return (
    <div className='group relative aspect-square overflow-hidden rounded-lg bg-muted'>
      {/* Media content - wrapped in Link for navigation */}
      <Link
        href={`/media/${item.id}`}
        className='absolute inset-0 w-full h-full cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2'
      >
        {isLoading && !isError && (
          <div className='absolute inset-0 flex items-center justify-center'>
            <div className='size-8 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent' />
          </div>
        )}

        {isError && (
          <div className='absolute inset-0 flex items-center justify-center text-muted-foreground text-sm'>
            Failed to load
          </div>
        )}

        {!isError && (
          <Image
            src={thumbnailUrl}
            alt=''
            fill
            unoptimized
            className={cn(
              'object-cover transition-transform duration-200',
              'group-hover:scale-105'
            )}
            sizes='(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw'
            onLoad={() => setIsLoading(false)}
            onError={() => {
              setIsLoading(false)
              setIsError(true)
            }}
          />
        )}

        {/* Video indicator */}
        {item.type === 'video' && !isLoading && !isError && (
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

        {/* Metadata overlay on hover */}
        {!isLoading && !isError && (displayDate || cameraInfo) && (
          <div className='absolute bottom-0 left-0 right-0 bg-linear-to-t from-black/70 to-transparent p-2 opacity-0 group-hover:opacity-100 transition-opacity'>
            <div className='flex items-center gap-2 text-xs text-white/90'>
              {displayDate && (
                <span className='flex items-center gap-1'>
                  <Calendar className='size-3' />
                  {displayDate}
                </span>
              )}
              {cameraInfo && (
                <span className='flex items-center gap-1 truncate'>
                  <Camera className='size-3' />
                  <span className='truncate'>{cameraInfo}</span>
                </span>
              )}
            </div>
          </div>
        )}
      </Link>

      {/* Action menu */}
      <div className='absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity z-10'>
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
