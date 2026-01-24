import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

interface RouteParams {
  params: Promise<{ id: string }>
}

// GET /api/media/[id]/thumbnail - Stream thumbnail image with caching
export async function GET(_request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(
      `${getApiBaseUrl()}/v1/media/${id}/thumbnail`,
      {
        headers: getApiHeaders()
      }
    )

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Thumbnail not found' },
        { status: response.status }
      )
    }

    const contentType =
      response.headers.get('content-type') ?? 'image/jpeg'
    const contentLength = response.headers.get('content-length')
    const cacheControl = response.headers.get('cache-control')

    const headers = new Headers({
      'Content-Type': contentType
    })

    if (contentLength) {
      headers.set('Content-Length', contentLength)
    }

    // Forward caching headers from backend
    if (cacheControl) {
      headers.set('Cache-Control', cacheControl)
    } else {
      // Default to aggressive caching for thumbnails
      headers.set('Cache-Control', 'public, max-age=31536000, immutable')
    }

    // Stream the response body
    return new NextResponse(response.body, {
      status: 200,
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
