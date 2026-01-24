import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

interface RouteParams {
  params: Promise<{ id: string }>
}

// GET /api/media/[id]/download - Stream media file
export async function GET(_request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(`${getApiBaseUrl()}/v1/media/${id}/download`, {
      headers: getApiHeaders()
    })

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Media not found' },
        { status: response.status }
      )
    }

    const contentType =
      response.headers.get('content-type') ?? 'application/octet-stream'
    const contentLength = response.headers.get('content-length')

    const headers = new Headers({
      'Content-Type': contentType
    })

    if (contentLength) {
      headers.set('Content-Length', contentLength)
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
