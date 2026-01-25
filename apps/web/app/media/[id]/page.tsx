'use client'

import { useState } from 'react'
import Image from 'next/image'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  ArrowLeft,
  Calendar,
  Camera,
  MapPin,
  FileImage,
  Clock,
  Aperture,
  Focus,
  Sun,
  Download,
  Loader2
} from 'lucide-react'
import { useMediaItem } from '@/hooks/use-media'
import { getMediaUrl } from '@/lib/media-api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { MediaItem } from '@/lib/types/media'

// Helper functions

function formatDate(dateString?: string): string | null {
  if (!dateString) return null
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString(undefined, {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch {
    return null
  }
}

function formatFileSize(bytes?: number): string | null {
  if (!bytes) return null
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`
}

function formatDuration(seconds?: number): string | null {
  if (!seconds) return null
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  if (mins > 0) {
    return `${mins}m ${secs}s`
  }
  return `${secs}s`
}

function formatCoordinates(lat?: number, lon?: number): string | null {
  if (lat === undefined || lon === undefined) return null
  const latDir = lat >= 0 ? 'N' : 'S'
  const lonDir = lon >= 0 ? 'E' : 'W'
  return `${Math.abs(lat).toFixed(6)}° ${latDir}, ${Math.abs(lon).toFixed(6)}° ${lonDir}`
}

// Sub-components

function MediaHeader({ id, filename }: { id: string; filename?: string }) {
  return (
    <header className='sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60 px-3'>
      <div className='container flex h-14 items-center gap-4'>
        <Link href='/'>
          <Button variant='ghost' size='sm'>
            <ArrowLeft className='size-4 mr-2' />
            Back
          </Button>
        </Link>
        <div className='flex-1' />
        <a href={getMediaUrl(id)} download={filename || `media-${id}`}>
          <Button variant='outline' size='sm'>
            <Download className='size-4 mr-2' />
            Download
          </Button>
        </a>
      </div>
    </header>
  )
}

function MediaDisplay({
  item,
  mediaUrl,
  isLoading,
  onLoad
}: {
  item: MediaItem
  mediaUrl: string
  isLoading: boolean
  onLoad: () => void
}) {
  return (
    <div className='relative flex items-center justify-center min-h-[400px] lg:min-h-[600px] bg-muted rounded-lg overflow-hidden'>
      {isLoading && (
        <div className='absolute inset-0 flex items-center justify-center'>
          <Loader2 className='size-8 animate-spin text-muted-foreground' />
        </div>
      )}

      {item.type === 'photo' && (
        <Image
          src={mediaUrl}
          alt={item.originalFilename || 'Photo'}
          width={item.width || 1920}
          height={item.height || 1080}
          className='max-w-full max-h-[80vh] w-auto h-auto object-contain'
          onLoad={onLoad}
          priority
        />
      )}

      {item.type === 'video' && (
        <video
          src={mediaUrl}
          controls
          autoPlay
          className='max-w-full max-h-[80vh]'
          onLoadedData={onLoad}
        >
          Your browser does not support the video tag.
        </video>
      )}
    </div>
  )
}

function MetadataPanel({ item }: { item: MediaItem }) {
  const displayDate = formatDate(item.takenAt) || formatDate(item.createdAt)
  const coordinates = formatCoordinates(item.latitude, item.longitude)
  const fileSize = formatFileSize(item.sizeBytes)
  const dimensions =
    item.width && item.height ? `${item.width} × ${item.height}` : null
  const duration = formatDuration(item.durationSec)

  return (
    <div className='space-y-3'>
      {displayDate && <DateCard date={displayDate} />}
      {(item.cameraMake || item.cameraModel) && (
        <CameraCard make={item.cameraMake} model={item.cameraModel} />
      )}
      {(item.iso || item.fNumber || item.exposureTime || item.focalLength) && (
        <SettingsCard item={item} />
      )}
      {coordinates && (
        <LocationCard
          coordinates={coordinates}
          lat={item.latitude!}
          lon={item.longitude!}
        />
      )}
      <FileCard
        item={item}
        dimensions={dimensions}
        fileSize={fileSize}
        duration={duration}
      />
    </div>
  )
}

function DateCard({ date }: { date: string }) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium flex items-center gap-2'>
          <Calendar className='size-4' />
          Date & Time
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className='text-sm text-muted-foreground'>{date}</p>
      </CardContent>
    </Card>
  )
}

function CameraCard({ make, model }: { make?: string; model?: string }) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium flex items-center gap-2'>
          <Camera className='size-4' />
          Camera
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-1'>
        {make && <p className='text-sm text-muted-foreground'>{make}</p>}
        {model && <p className='text-sm font-medium'>{model}</p>}
      </CardContent>
    </Card>
  )
}

function SettingsCard({ item }: { item: MediaItem }) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium flex items-center gap-2'>
          <Aperture className='size-4' />
          Camera Settings
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='grid grid-cols-2 gap-2 text-sm'>
          {item.fNumber && (
            <div className='flex items-center gap-1.5'>
              <Aperture className='size-3.5 text-muted-foreground' />
              <span>f/{item.fNumber.toFixed(1)}</span>
            </div>
          )}
          {item.exposureTime && (
            <div className='flex items-center gap-1.5'>
              <Clock className='size-3.5 text-muted-foreground' />
              <span>{item.exposureTime}s</span>
            </div>
          )}
          {item.iso && (
            <div className='flex items-center gap-1.5'>
              <Sun className='size-3.5 text-muted-foreground' />
              <span>ISO {item.iso}</span>
            </div>
          )}
          {item.focalLength && (
            <div className='flex items-center gap-1.5'>
              <Focus className='size-3.5 text-muted-foreground' />
              <span>{item.focalLength.toFixed(0)}mm</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function LocationCard({
  coordinates,
  lat,
  lon
}: {
  coordinates: string
  lat: number
  lon: number
}) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium flex items-center gap-2'>
          <MapPin className='size-4' />
          Location
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className='text-sm text-muted-foreground font-mono'>{coordinates}</p>
        <a
          href={`https://www.google.com/maps?q=${lat},${lon}`}
          target='_blank'
          rel='noopener noreferrer'
          className='text-xs text-primary hover:underline mt-1 inline-block'
        >
          View on Google Maps
        </a>
      </CardContent>
    </Card>
  )
}

function FileCard({
  item,
  dimensions,
  fileSize,
  duration
}: {
  item: MediaItem
  dimensions: string | null
  fileSize: string | null
  duration: string | null
}) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-sm font-medium flex items-center gap-2'>
          <FileImage className='size-4' />
          File Details
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-1 text-sm'>
        {item.originalFilename && (
          <p className='text-muted-foreground truncate'>
            {item.originalFilename}
          </p>
        )}
        <div className='flex flex-wrap gap-x-4 gap-y-1 text-muted-foreground'>
          <span className='capitalize'>{item.type}</span>
          {dimensions && <span>{dimensions}</span>}
          {fileSize && <span>{fileSize}</span>}
          {duration && <span>{duration}</span>}
        </div>
        {item.mimeType && (
          <p className='text-xs text-muted-foreground/70 font-mono'>
            {item.mimeType}
          </p>
        )}
      </CardContent>
    </Card>
  )
}

// Main page component

export default function MediaDetailPage() {
  const params = useParams()
  const id = params.id as string
  const { data: item, isLoading, error } = useMediaItem(id)
  const [isMediaLoading, setIsMediaLoading] = useState(true)

  const mediaUrl = getMediaUrl(id)

  if (isLoading) {
    return (
      <div className='min-h-screen flex items-center justify-center'>
        <Loader2 className='size-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  if (error || !item) {
    return (
      <div className='min-h-screen flex flex-col items-center justify-center gap-4'>
        <p className='text-lg text-muted-foreground'>Media not found</p>
        <Link href='/'>
          <Button variant='outline'>
            <ArrowLeft className='size-4 mr-2' />
            Back to Gallery
          </Button>
        </Link>
      </div>
    )
  }

  return (
    <div className='min-h-screen bg-background'>
      <MediaHeader id={id} filename={item.originalFilename} />
      <main className='container py-6 mx-auto'>
        <div className='grid gap-6 lg:grid-cols-[1fr_320px] px-3 sm:px-0'>
          <MediaDisplay
            item={item}
            mediaUrl={mediaUrl}
            isLoading={isMediaLoading}
            onLoad={() => setIsMediaLoading(false)}
          />
          <MetadataPanel item={item} />
        </div>
      </main>
    </div>
  )
}
