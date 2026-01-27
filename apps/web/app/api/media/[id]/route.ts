import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

interface RouteParams {
  params: Promise<{ id: string }>
}

// GET /api/media/[id] - Get media metadata
export async function GET(request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(`${getApiBaseUrl()}/media/${id}`, {
      headers: getApiHeaders()
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to API' },
      { status: 502 }
    )
  }
}

// DELETE /api/media/[id] - Delete media
export async function DELETE(request: NextRequest, { params }: RouteParams) {
  const { id } = await params

  try {
    const response = await fetch(`${getApiBaseUrl()}/media/${id}`, {
      method: 'DELETE',
      headers: getApiHeaders()
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to API' },
      { status: 502 }
    )
  }
}
