import { MediaGallery } from '@/components/media-gallery'
import { Header } from '@/components/header'
import { validateUser } from '@/lib/validate-user'

export default async function Home() {
  // Validate user is in household - redirects if not authorized
  await validateUser()

  return (
    <div className='min-h-svh bg-background'>
      <Header />
      <main className='p-4 sm:p-8'>
        <MediaGallery />
      </main>
    </div>
  )
}
