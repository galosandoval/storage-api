'use client'

import { Loader2 } from 'lucide-react'
import type { MediaItem } from '@/lib/types/media'
import { MediaCard } from '@/components/media-card'

interface MediaGridProps {
  items: MediaItem[]
  isLoading: boolean
  isLoadingMore: boolean
  hasMore: boolean
  onItemDelete: (item: MediaItem) => void
  sentinelRef: (node: HTMLElement | null) => void
}

export function MediaGrid({
  items,
  isLoading,
  isLoadingMore,
  hasMore,
  onItemDelete,
  sentinelRef
}: MediaGridProps) {
  if (isLoading) {
    return (
      <div className='flex items-center justify-center py-20'>
        <Loader2 className='size-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className='flex flex-col items-center justify-center py-20 text-muted-foreground'>
        <p className='text-lg'>No media found</p>
        <p className='text-sm mt-1'>
          Upload some photos or videos to get started
        </p>
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <div className='grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-2 sm:gap-3'>
        {items.map((item) => (
          <MediaCard
            key={item.id}
            item={item}
            onDelete={() => onItemDelete(item)}
          />
        ))}
      </div>

      {/* Infinite scroll sentinel */}
      {hasMore && (
        <div
          ref={sentinelRef}
          className='flex items-center justify-center py-8'
        >
          {isLoadingMore && (
            <Loader2 className='size-6 animate-spin text-muted-foreground' />
          )}
        </div>
      )}
    </div>
  )
}
