'use client'

import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { deleteMedia } from '@/lib/media-api'
import type { MediaItem } from '@/lib/types/media'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'

interface DeleteDialogProps {
  item: MediaItem | null
  isOpen: boolean
  onClose: () => void
  onDeleted: (id: string) => void
}

export function DeleteDialog({
  item,
  isOpen,
  onClose,
  onDeleted
}: DeleteDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleDelete = async () => {
    if (!item) return

    try {
      setIsDeleting(true)
      setError(null)
      await deleteMedia(item.id)
      onDeleted(item.id)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete')
    } finally {
      setIsDeleting(false)
    }
  }

  const handleOpenChange = (open: boolean) => {
    if (!open && !isDeleting) {
      onClose()
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            Delete {item?.type === 'video' ? 'video' : 'photo'}?
          </DialogTitle>
          <DialogDescription>
            This action cannot be undone. This will permanently delete this{' '}
            {item?.type === 'video' ? 'video' : 'photo'} from your storage.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className='rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive'>
            {error}
          </div>
        )}

        <DialogFooter>
          <Button variant='outline' onClick={onClose} disabled={isDeleting}>
            Cancel
          </Button>
          <Button
            variant='destructive'
            onClick={handleDelete}
            disabled={isDeleting}
          >
            {isDeleting && <Loader2 className='size-4 mr-2 animate-spin' />}
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
