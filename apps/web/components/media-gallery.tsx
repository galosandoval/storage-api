'use client'

import { useState, useCallback } from 'react'
import { RefreshCw } from 'lucide-react'
import { useMedia } from '@/hooks/use-media'
import type { MediaItem, MediaTypeFilter } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import { MediaGrid } from '@/components/media-grid'
import { MediaLightbox } from '@/components/media-lightbox'
import { UploadDropzone } from '@/components/upload-dropzone'
import { TypeFilter } from '@/components/type-filter'
import { DeleteDialog } from '@/components/delete-dialog'

export function MediaGallery() {
  const [typeFilter, setTypeFilter] = useState<MediaTypeFilter>('all')
  const [lightboxIndex, setLightboxIndex] = useState(-1)
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

  const handleItemClick = useCallback((_item: MediaItem, index: number) => {
    setLightboxIndex(index)
  }, [])

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
      // If viewing this item in lightbox, close it
      if (deleteItem?.id === id && lightboxIndex >= 0) {
        setLightboxIndex(-1)
      }
    },
    [removeItem, deleteItem, lightboxIndex]
  )

  return (
    <div className='w-full max-w-6xl mx-auto space-y-6'>
      {/* Header */}
      <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4'>
        <h1 className='text-2xl font-semibold tracking-tight'>Media Gallery</h1>
        <div className='flex items-center gap-2'>
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
        onItemClick={handleItemClick}
        onItemDelete={handleItemDelete}
        sentinelRef={sentinelRef}
      />

      {/* Lightbox */}
      <MediaLightbox
        items={items}
        currentIndex={lightboxIndex}
        isOpen={lightboxIndex >= 0}
        onClose={() => setLightboxIndex(-1)}
        onIndexChange={setLightboxIndex}
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
