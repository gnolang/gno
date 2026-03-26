export const PATH_SEPARATOR = '/'

export const abspath = (fpath: string, root = '') =>
  fpath[0] === PATH_SEPARATOR ? fpath : root + PATH_SEPARATOR + fpath

const dirNameFromIndex = (str: string, i: number) => (i <= 0 ? PATH_SEPARATOR : abspath(str.slice(0, i)))

export const basename = (fpath: string) => fpath.slice(fpath.lastIndexOf(PATH_SEPARATOR) + 1)
export const dirname = (fpath: string) => dirNameFromIndex(fpath, fpath.lastIndexOf(PATH_SEPARATOR))
export const splitPath = (fpath: string) => {
  const i = fpath.lastIndexOf(PATH_SEPARATOR)
  return [dirNameFromIndex(fpath, i), fpath.slice(i + 1)]
}

export const trimPrefix = (str: string, pfx: string) => {
  if (str.slice(0, pfx.length) === pfx) {
    return str.slice(pfx.length)
  }

  return str
}
