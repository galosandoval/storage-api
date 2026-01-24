'use client'

import { useCallback, useEffect, useRef } from 'react'
import {
  useInfiniteQuery,
  useQueryClient,
  useQuery
} from '@tanstack/react-query'
import { listMedia, getMedia, getThumbnailBlob } from '@/lib/media-api'
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
  const queryClient = useQueryClient()

  const observerRef = useRef<IntersectionObserver | null>(null)
  const sentinelNodeRef = useRef<HTMLElement | null>(null)

  const {
    data,
    isLoading,
    isFetchingNextPage,
    error,
    hasNextPage,
    fetchNextPage,
    refetch
  } = useInfiniteQuery({
    queryKey: ['media', typeFilter, pageSize],
    queryFn: async ({ pageParam = 1 }) => {
      return listMedia(pageParam, pageSize, typeFilter)
    },
    getNextPageParam: (lastPage) => {
      const hasMore = lastPage.page * lastPage.pageSize < lastPage.totalCount
      return hasMore ? lastPage.page + 1 : undefined
    },
    initialPageParam: 1
  })

  // Flatten pages into items array
  const items = data?.pages.flatMap((page) => page.items) ?? []

  // Load more callback
  const loadMore = useCallback(() => {
    if (!isFetchingNextPage && hasNextPage) {
      fetchNextPage()
    }
  }, [isFetchingNextPage, hasNextPage, fetchNextPage])

  // Refresh callback
  const refresh = useCallback(() => {
    refetch()
  }, [refetch])

  // Prepend a new item (after upload) - optimistic update
  const prependItem = useCallback(
    (item: MediaItem) => {
      queryClient.setQueryData(
        ['media', typeFilter, pageSize],
        (oldData: typeof data) => {
          if (!oldData) return oldData
          return {
            ...oldData,
            pages: oldData.pages.map((page, index) => {
              if (index === 0) {
                return {
                  ...page,
                  items: [item, ...page.items],
                  totalCount: page.totalCount + 1
                }
              }
              return page
            })
          }
        }
      )
    },
    [queryClient, typeFilter, pageSize]
  )

  // Remove an item (after delete) - optimistic update
  const removeItem = useCallback(
    (id: string) => {
      queryClient.setQueryData(
        ['media', typeFilter, pageSize],
        (oldData: typeof data) => {
          if (!oldData) return oldData
          return {
            ...oldData,
            pages: oldData.pages.map((page) => ({
              ...page,
              items: page.items.filter((item) => item.id !== id),
              totalCount: page.totalCount - 1
            }))
          }
        }
      )
    },
    [queryClient, typeFilter, pageSize]
  )

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
            hasNextPage &&
            !isFetchingNextPage &&
            !isLoading
          ) {
            loadMore()
          }
        },
        { rootMargin: '200px' }
      )

      observerRef.current.observe(node)
    },
    [hasNextPage, isFetchingNextPage, isLoading, loadMore]
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
    isLoadingMore: isFetchingNextPage,
    error: error instanceof Error ? error.message : null,
    hasMore: hasNextPage ?? false,
    loadMore,
    refresh,
    prependItem,
    removeItem,
    sentinelRef
  }
}

// Hook for fetching a single media item
export function useMediaItem(id: string) {
  return useQuery({
    queryKey: ['media', 'item', id],
    queryFn: () => getMedia(id),
    enabled: !!id
  })
}

// Hook for fetching a thumbnail blob URL
export function useThumbnail(id: string) {
  return useQuery({
    queryKey: ['thumbnail', id],
    queryFn: () => getThumbnailBlob(id),
    enabled: !!id,
    staleTime: Infinity, // Thumbnails don't change
    gcTime: 30 * 60 * 1000 // Keep in cache for 30 minutes
  })
}
