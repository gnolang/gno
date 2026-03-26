const moduleChunksRegEx = /^([a-zA-Z._-]+)\/(p|r)\/([^\s/]+)\/(\S+)$/i

/**
 * Splits module path into chunks.
 *
 * @example `gno.land/p/foo/bar`
 * @param pkgPath Package path string
 */
export const parsePkgPath = (pkgPath: string) => {
  const matches = moduleChunksRegEx.exec(pkgPath)
  if (!matches) {
    return {
      path: pkgPath,
    }
  }

  const [, domain, type, namespace, path] = matches
  return { domain, type, namespace, path }
}

export function isValidPkgPath(
  pkgPath: string,
  optionals: {
    domain?: boolean
    type?: boolean
  } = {},
) {
  const rePkgPath = '[a-zA-Z0-9_]+(/[a-zA-Z0-9_]+)+$'
  let reDomain = '^(gno.land/)'
  let reType = '([rp]/)'

  if (optionals.domain === false) reDomain += '?'
  if (optionals.type === false) reType += '?'

  const re = new RegExp(reDomain + reType + rePkgPath)

  return re.test(pkgPath)
}
