import { SignOutButton, UserButton } from '@clerk/nextjs'
import { LogOut } from 'lucide-react'
import { MediaGallery } from '@/components/media-gallery'
import { Button } from '@/components/ui/button'

export default function Home() {
  return (
    <div className='min-h-svh bg-background'>
      <header className='flex items-center justify-between p-4 sm:px-8 border-b'>
        <h1 className='text-xl font-semibold'>Laulo Media</h1>
        <div className='flex items-center gap-3'>
          <SignOutButton>
            <Button variant='outline' size='sm'>
              <LogOut className='size-4 mr-2' />
              Sign out
            </Button>
          </SignOutButton>
          <UserButton />
        </div>
      </header>
      <main className='p-4 sm:p-8'>
        <MediaGallery />
      </main>
    </div>
  )
}
