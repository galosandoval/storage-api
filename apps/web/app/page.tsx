import { MediaGallery } from '@/components/media-gallery'
import { getApiBaseUrl } from '@/lib/api-config'

export default function Home() {
  console.log('Home')
  console.log('getApiBaseUrl', getApiBaseUrl())
  return (
    <div className='min-h-svh bg-background p-4 sm:p-8'>
      <MediaGallery />
    </div>
  )
}
