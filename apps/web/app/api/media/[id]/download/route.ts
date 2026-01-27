import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

interface RouteParams {
  params: Promise<{ id: string }>
}

// GET /api/media/[id]/download - Download media file
export async function GET(request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(`${getApiBaseUrl()}/media/${id}/download`, {
      headers: await getApiHeaders()
    })

    if (!response.ok) {
      const data = await response.json()
      return NextResponse.json(data, { status: response.status })
    }

    // Stream the file response
    const headers = new Headers()
    response.headers.forEach((value, key) => {
      if (
        key.toLowerCase() === 'content-type' ||
        key.toLowerCase() === 'content-disposition' ||
        key.toLowerCase() === 'content-length'
      ) {
        headers.set(key, value)
      }
    })

    return new NextResponse(response.body, {
      status: response.status,
      headers
    })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to API' },
      { status: 502 }
    )
  }
}
