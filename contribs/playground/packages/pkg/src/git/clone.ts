import * as BrowserFS from 'browserfs'
import { type FSModule } from 'browserfs/dist/node/core/FS'
import { clone } from 'isomorphic-git'
import http from 'isomorphic-git/http/web'

import { BackendInmemory } from '../xbfs'

export interface PlainFile {
  path: string
  content: string
}

function getAllValidFiles(root: string, fs: FSModule) {
  const result: PlainFile[] = []
  const validExtensions = ['.go', '.gno', '.mod', '.toml', '.yaml', '.yml', '.md']

  const walk = (dir: string) => {
    fs.readdirSync(dir).forEach((file: string) => {
      const fullPath = `${dir}/${file}`

      if (file.startsWith('.')) return

      if (fs.statSync(fullPath).isDirectory()) {
        walk(fullPath)
      } else {
        if (!validExtensions.some((ext) => file.endsWith(ext))) return

        result.push({
          path: fullPath,
          content: fs.readFileSync(fullPath, 'utf8'),
        })
      }
    })
  }

  walk(root)
  return result
}

/**
 * Clones a git repository to a in-memory filesystem
 * and returns all the valid files in the repository.
 */
export async function cloneGitRepository(url: string, branch?: string) {
  const inMemory = await BackendInmemory()
  BrowserFS.initialize(inMemory)

  const fs = BrowserFS.BFSRequire('fs')
  const corsProxy = import.meta.env.VITE_BACKEND_GIT_PROXY_URL
  const dir = '.'

  await clone({
    fs,
    http,
    dir,
    corsProxy,
    url,
    singleBranch: true,
    ref: branch,
    depth: 1,
  })

  return getAllValidFiles(dir, fs)
}
