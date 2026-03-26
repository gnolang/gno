import { type PlainFile } from '../types'

// TODO: Re-integrate cloud session sharing when a backend is available

interface SessionPayload {
  description: string
  files: PlainFile[]
}

export async function createSession(_payload: SessionPayload): Promise<{ uri: string }> {
  console.warn('Cloud session sharing not available in this deployment')
  return { uri: '' }
}

export async function getSession(_uri: string): Promise<SessionPayload> {
  console.warn('Cloud session sharing not available in this deployment')
  return { description: '', files: [] }
}
