'use client'

import { Image, Video, LayoutGrid } from 'lucide-react'
import { useTranslation } from '@/hooks/use-translations'
import type { MediaTypeFilter } from '@/lib/types/media'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

interface TypeFilterProps {
  value: MediaTypeFilter
  onChange: (value: MediaTypeFilter) => void
}

export function TypeFilter({ value, onChange }: TypeFilterProps) {
  const { t } = useTranslation('filters')

  return (
    <ToggleGroup
      type='single'
      value={value}
      onValueChange={(val) => {
        if (val) {
          onChange(val as MediaTypeFilter)
        }
      }}
      className='justify-start'
    >
      <ToggleGroupItem value='all' aria-label={t('showAll')}>
        <LayoutGrid className='size-4 mr-2' />
        {t('all')}
      </ToggleGroupItem>
      <ToggleGroupItem value='photo' aria-label={t('showPhotos')}>
        <Image className='size-4 mr-2' />
        {t('photos')}
      </ToggleGroupItem>
      <ToggleGroupItem value='video' aria-label={t('showVideos')}>
        <Video className='size-4 mr-2' />
        {t('videos')}
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
