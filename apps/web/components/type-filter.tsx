'use client'

import { Image, Video, LayoutGrid } from 'lucide-react'
import type { MediaTypeFilter } from '@/lib/types/media'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

interface TypeFilterProps {
  value: MediaTypeFilter
  onChange: (value: MediaTypeFilter) => void
}

export function TypeFilter({ value, onChange }: TypeFilterProps) {
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
      <ToggleGroupItem value='all' aria-label='Show all'>
        <LayoutGrid className='size-4 mr-2' />
        All
      </ToggleGroupItem>
      <ToggleGroupItem value='photo' aria-label='Show photos'>
        <Image className='size-4 mr-2' />
        Photos
      </ToggleGroupItem>
      <ToggleGroupItem value='video' aria-label='Show videos'>
        <Video className='size-4 mr-2' />
        Videos
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
