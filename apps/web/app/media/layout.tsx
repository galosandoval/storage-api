import { validateUser } from '@/lib/validate-user'

export default async function MediaLayout({
  children
}: {
  children: React.ReactNode
}) {
  // Validate user is in household - redirects if not authorized
  await validateUser()

  return <>{children}</>
}
