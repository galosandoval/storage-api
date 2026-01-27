'use client'

import { useState, useCallback } from 'react'
import { RefreshCw } from 'lucide-react'
import { useMedia } from '@/hooks/use-media'
import type { MediaItem, MediaTypeFilter } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import { MediaGrid } from '@/components/media-grid'
import { UploadDropzone } from '@/components/upload-dropzone'
import { TypeFilter } from '@/components/type-filter'
import { DeleteDialog } from '@/components/delete-dialog'

export function MediaGallery() {
  const [typeFilter, setTypeFilter] = useState<MediaTypeFilter>('all')
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
  } = useMedia({ typeFilter })

  const handleItemDelete = useCallback((item: MediaItem) => {
    setDeleteItem(item)
  }, [])

  const handleUploadComplete = useCallback(
    (item: MediaItem) => {
      // Only prepend if it matches the current filter
      if (typeFilter === 'all' || typeFilter === item.type) {
        prependItem(item)
      }
    },
    [typeFilter, prependItem]
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
      <div className='flex items-center justify-end gap-2'>
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
