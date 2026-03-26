const VITE_GNOLAND_URL = 'https://gno.land'
const VITE_REALM_PATH_PREFIX = 'gno.land/'

export const urls = {
  gnoLand: VITE_GNOLAND_URL,
}

/**
 * Helpers and constants to work with Gno realm and package paths.
 */
export const packagePaths = {
  realmNamespace: 'r',
  packageNamespace: 'p',
  pathPrefix: VITE_REALM_PATH_PREFIX,

  isRealmPath(path: string) {
    if (path.startsWith(this.pathPrefix)) {
      path = path.substring(this.pathPrefix.length)
    }

    return path.startsWith(`${this.realmNamespace}/`)
  },
}
