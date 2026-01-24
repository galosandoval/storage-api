'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { listMedia } from '@/lib/media-api'
import type { MediaItem, MediaTypeFilter } from '@/lib/types/media'

interface UseMediaOptions {
  pageSize?: number
  typeFilter?: MediaTypeFilter
}

interface UseMediaReturn {
  items: MediaItem[]
  isLoading: boolean
  isLoadingMore: boolean
  error: string | null
  hasMore: boolean
  loadMore: () => void
  refresh: () => void
  prependItem: (item: MediaItem) => void
  removeItem: (id: string) => void
  sentinelRef: (node: HTMLElement | null) => void
}

export function useMedia(options: UseMediaOptions = {}): UseMediaReturn {
  const { pageSize = 20, typeFilter = 'all' } = options

  const [items, setItems] = useState<MediaItem[]>([])
  const [page, setPage] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)

  const observerRef = useRef<IntersectionObserver | null>(null)
  const sentinelNodeRef = useRef<HTMLElement | null>(null)

  // Fetch media items
  const fetchMedia = useCallback(
    async (pageNum: number, append: boolean) => {
      try {
        if (append) {
          setIsLoadingMore(true)
        } else {
          setIsLoading(true)
        }
        setError(null)

        const response = await listMedia(pageNum, pageSize, typeFilter)

        setItems((prev) => {
          if (append) {
            // Merge avoiding duplicates
            const existingIds = new Set(prev.map((item) => item.id))
            const newItems = response.items.filter(
              (item) => !existingIds.has(item.id)
            )
            return [...prev, ...newItems]
          }
          return response.items
        })

        setHasMore(response.page * response.pageSize < response.totalCount)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load media')
      } finally {
        setIsLoading(false)
        setIsLoadingMore(false)
      }
    },
    [pageSize, typeFilter]
  )

  // Initial load and filter change
  useEffect(() => {
    setPage(1)
    setItems([])
    setHasMore(true)
    fetchMedia(1, false)
  }, [fetchMedia])

  // Load more
  const loadMore = useCallback(() => {
    if (isLoadingMore || !hasMore) return
    const nextPage = page + 1
    setPage(nextPage)
    fetchMedia(nextPage, true)
  }, [page, isLoadingMore, hasMore, fetchMedia])

  // Refresh
  const refresh = useCallback(() => {
    setPage(1)
    setItems([])
    setHasMore(true)
    fetchMedia(1, false)
  }, [fetchMedia])

  // Prepend a new item (after upload)
  const prependItem = useCallback((item: MediaItem) => {
    setItems((prev) => [item, ...prev])
  }, [])

  // Remove an item (after delete)
  const removeItem = useCallback((id: string) => {
    setItems((prev) => prev.filter((item) => item.id !== id))
  }, [])

  // Sentinel ref callback for IntersectionObserver
  const sentinelRef = useCallback(
    (node: HTMLElement | null) => {
      if (observerRef.current) {
        observerRef.current.disconnect()
      }

      if (!node) {
        sentinelNodeRef.current = null
        return
      }

      sentinelNodeRef.current = node

      observerRef.current = new IntersectionObserver(
        (entries) => {
          if (
            entries[0]?.isIntersecting &&
            hasMore &&
            !isLoadingMore &&
            !isLoading
          ) {
            loadMore()
          }
        },
        { rootMargin: '200px' }
      )

      observerRef.current.observe(node)
    },
    [hasMore, isLoadingMore, isLoading, loadMore]
  )

  // Cleanup observer on unmount
  useEffect(() => {
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect()
      }
    }
  }, [])

  return {
    items,
    isLoading,
    isLoadingMore,
    error,
    hasMore,
    loadMore,
    refresh,
    prependItem,
    removeItem,
    sentinelRef
  }
}
