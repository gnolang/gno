import { describe, expect, it } from 'vitest'

import { buildPlaygroundGithubUrl, parseGithubUrl } from './github'

describe('parseGithubUrl', () => {
  it('parses repository root', () => {
    expect(parseGithubUrl('https://github.com/foo/bar')).toEqual({ owner: 'foo', repo: 'bar' })
  })

  it('parses tree URLs with branch', () => {
    expect(parseGithubUrl('https://github.com/foo/bar/tree/dev')).toEqual({ owner: 'foo', repo: 'bar', branch: 'dev' })
  })

  it('parses blob URLs with branch and file', () => {
    expect(parseGithubUrl('https://github.com/foo/bar/blob/main/path/to/file.gno')).toEqual({
      owner: 'foo',
      repo: 'bar',
      branch: 'main',
      filePath: 'path/to/file.gno',
    })
  })

  it('accepts URLs without protocol', () => {
    expect(parseGithubUrl('github.com/foo/bar')).toEqual({ owner: 'foo', repo: 'bar' })
  })

  it('returns null for missing owner or repo', () => {
    expect(parseGithubUrl('https://github.com/foo')).toBeNull()
  })

  it('returns null for unsupported hosts', () => {
    expect(parseGithubUrl('https://gitlab.com/foo/bar')).toBeNull()
  })
})

describe('buildPlaygroundGithubUrl', () => {
  it('builds repo URL', () => {
    expect(buildPlaygroundGithubUrl('https://play.gno.land', { owner: 'foo', repo: 'bar' })).toBe(
      'https://play.gno.land/github/foo/bar',
    )
  })

  it('builds file URL with query params', () => {
    expect(
      buildPlaygroundGithubUrl('https://play.gno.land', {
        owner: 'foo',
        repo: 'bar',
        branch: 'dev',
        filePath: 'dir/file.gno',
      }),
    ).toBe('https://play.gno.land/github/foo/bar?file=dir%2Ffile.gno&branch=dev')
  })
})
