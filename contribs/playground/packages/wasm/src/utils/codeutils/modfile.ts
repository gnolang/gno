import { File } from '../../go/types'

export const GNOMOD_TOML = 'gnomod.toml'
export const GNOMOD_LEGACY = 'gno.mod'
export const GNOWORK_TOML = 'gnowork.toml'

// TOML format: module = "path"
const moduleTomlRegEx = /^module\s*=\s*"([^"]+)"/m
// Legacy format: module path
const moduleLegacyRegEx = /^module\s+"([^"]+)"/m

/**
 * Extracts Gno module name from gnomod.toml or gno.mod file contents.
 *
 * @param src Source file content
 */
export const extractModuleName = (src: string) => {
  // Try TOML format first
  let matches = moduleTomlRegEx.exec(src)

  // Fallback to legacy format
  if (!matches) {
    matches = moduleLegacyRegEx.exec(src)
  }

  if (!matches) {
    return
  }

  return matches[1]
}

type FileSet = Record<string, File>

/**
 * If a set of files does not contain a gnowork.toml file, adds an empty one.
 * This ensures that the GnoVM treats the files as part of a workspace.
 * @see https://docs.gno.land/resources/configuring-gno-projects/#workspaces-with-gnoworktoml
 */
export const withGnoWorkFile = (files: FileSet): FileSet => {
  if (files[GNOWORK_TOML]) {
    return files
  }

  return {
    ...files,
    [GNOWORK_TOML]: {
      content: '',
      path: GNOWORK_TOML,
    },
  }
}

export interface ModulePathSegments {
  kind?: string
  namespace?: string
  path: string
}
