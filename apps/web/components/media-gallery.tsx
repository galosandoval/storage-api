'use client'

import { useState, useCallback } from 'react'
import { RefreshCw } from 'lucide-react'
import { useMedia } from '@/hooks/use-media'
import type { MediaItem, MediaTypeFilter, VisibilityFilter as VisibilityFilterType } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import { MediaGrid } from '@/components/media-grid'
import { UploadDropzone } from '@/components/upload-dropzone'
import { TypeFilter } from '@/components/type-filter'
import { VisibilityFilter } from '@/components/visibility-filter'
import { DeleteDialog } from '@/components/delete-dialog'

export function MediaGallery() {
  const [typeFilter, setTypeFilter] = useState<MediaTypeFilter>('all')
  const [visibilityFilter, setVisibilityFilter] = useState<VisibilityFilterType>('all')
  const [deleteItem, setDeleteItem] = useState<MediaItem | null>(null)

  const {
    items,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    refresh,
    prependItem,
    removeItem,
    sentinelRef
  } = useMedia({ typeFilter, visibilityFilter })

  const handleItemDelete = useCallback((item: MediaItem) => {
    setDeleteItem(item)
  }, [])

  const handleUploadComplete = useCallback(
    (item: MediaItem) => {
      // Only prepend if it matches the current filters
      const matchesType = typeFilter === 'all' || typeFilter === item.type
      const matchesVisibility =
        visibilityFilter === 'all' ||
        visibilityFilter === 'mine' ||
        (visibilityFilter === 'public' && !item.isPrivate)
      if (matchesType && matchesVisibility) {
        prependItem(item)
      }
    },
    [typeFilter, visibilityFilter, prependItem]
  )

  const handleDeleted = useCallback(
    (id: string) => {
      removeItem(id)
    },
    [removeItem]
  )

  return (
    <div className='w-full max-w-6xl mx-auto space-y-6'>
      {/* Controls */}
      <div className='flex flex-wrap items-center justify-end gap-2'>
        <VisibilityFilter value={visibilityFilter} onChange={setVisibilityFilter} />
        <TypeFilter value={typeFilter} onChange={setTypeFilter} />
        <Button
          variant='outline'
          size='icon'
          onClick={refresh}
          title='Refresh'
        >
          <RefreshCw className='size-4' />
        </Button>
      </div>

      {/* Upload dropzone */}
      <UploadDropzone onUploadComplete={handleUploadComplete} />

      {/* Error state */}
      {error && (
        <div className='rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive'>
          {error}
        </div>
      )}

      {/* Media grid */}
      <MediaGrid
        items={items}
        isLoading={isLoading}
        isLoadingMore={isLoadingMore}
        hasMore={hasMore}
        onItemDelete={handleItemDelete}
        sentinelRef={sentinelRef}
      />

      {/* Delete confirmation dialog */}
      <DeleteDialog
        item={deleteItem}
        isOpen={deleteItem !== null}
        onClose={() => setDeleteItem(null)}
        onDeleted={handleDeleted}
      />
    </div>
  )
}
