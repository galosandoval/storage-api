'use client'

import { useEffect, useState } from 'react'
import Lightbox, { type Slide } from 'yet-another-react-lightbox'
import Video from 'yet-another-react-lightbox/plugins/video'
import Zoom from 'yet-another-react-lightbox/plugins/zoom'
import 'yet-another-react-lightbox/styles.css'

import { getMediaBlob } from '@/lib/media-api'
import type { MediaItem } from '@/lib/types/media'

interface MediaLightboxProps {
  items: MediaItem[]
  currentIndex: number
  isOpen: boolean
  onClose: () => void
  onIndexChange: (index: number) => void
}

export function MediaLightbox({
  items,
  currentIndex,
  isOpen,
  onClose,
  onIndexChange
}: MediaLightboxProps) {
  const [slides, setSlides] = useState<Slide[]>([])
  const [loadedIds, setLoadedIds] = useState<Set<string>>(new Set())

  // Load blob URLs for visible items
  useEffect(() => {
    if (!isOpen) return

    async function loadSlides() {
      const newSlides: Slide[] = []

      for (const item of items) {
        if (loadedIds.has(item.id)) {
          // Find existing slide by index
          const existingIndex = items.findIndex((i) => i.id === item.id)
          if (existingIndex >= 0 && slides[existingIndex]) {
            newSlides.push(slides[existingIndex])
            continue
          }
        }

        try {
          const blobUrl = await getMediaBlob(item.id)

          if (item.type === 'photo') {
            newSlides.push({
              src: blobUrl,
              width: item.width,
              height: item.height
            })
          } else {
            newSlides.push({
              type: 'video',
              width: item.width,
              height: item.height,
              sources: [
                {
                  src: blobUrl,
                  type: item.mimeType || 'video/mp4'
                }
              ]
            })
          }

          setLoadedIds((prev) => new Set(prev).add(item.id))
        } catch {
          // Create placeholder for failed loads
          newSlides.push({
            src: ''
          })
        }
      }

      setSlides(newSlides)
    }

    loadSlides()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, items])

  // Cleanup blob URLs on unmount
  useEffect(() => {
    return () => {
      slides.forEach((slide) => {
        if ('src' in slide && slide.src && slide.src.startsWith('blob:')) {
          URL.revokeObjectURL(slide.src)
        }
        if ('sources' in slide && slide.sources) {
          slide.sources.forEach((source) => {
            if (source.src.startsWith('blob:')) {
              URL.revokeObjectURL(source.src)
            }
          })
        }
      })
    }
  }, [slides])

  if (!isOpen || slides.length === 0) {
    return null
  }

  return (
    <Lightbox
      open={isOpen}
      close={onClose}
      index={currentIndex}
      slides={slides}
      on={{
        view: ({ index }) => onIndexChange(index)
      }}
      plugins={[Video, Zoom]}
      video={{
        controls: true,
        autoPlay: true
      }}
      zoom={{
        maxZoomPixelRatio: 3,
        scrollToZoom: true
      }}
      carousel={{
        finite: false,
        preload: 2
      }}
      controller={{
        closeOnBackdropClick: true
      }}
    />
  )
}
