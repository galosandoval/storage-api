import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

interface RouteParams {
  params: Promise<{ id: string }>
}

// GET /api/media/[id]/thumbnail - Get media thumbnail
export async function GET(request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(`${getApiBaseUrl()}/media/${id}/thumbnail`, {
      headers: await getApiHeaders()
    })

    if (!response.ok) {
      const data = await response.json()
      return NextResponse.json(data, { status: response.status })
    }

    // Stream the image response
    const headers = new Headers()
    response.headers.forEach((value, key) => {
      if (
        key.toLowerCase() === 'content-type' ||
        key.toLowerCase() === 'content-length' ||
        key.toLowerCase() === 'cache-control'
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
