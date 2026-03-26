import { extractModuleName, GNOMOD_LEGACY, GNOMOD_TOML } from './modfile'

// Value type is mixed to share code between store and upcoming gopls module which have different types
type Workspace = Record<string, { content: string } | string>

/**
 * Returns whether file name corresponds to Gno source or module file.
 */
export const isGnoFile = (filename: string) => filename.match(/(\.gno|gnomod\.toml|gno\.mod)$/)

/**
 * Extracts Gno package name statement from source text.
 * @param src Program source
 */
export const parsePackageName = (src: string) => {
  const match = /^package\s(\w+)/m.exec(src)
  return match?.[1]
}

const getWorkspaceFile = (ws: Workspace, fname: string) => {
  const v = ws[fname]
  if (typeof v === 'string') {
    return v
  }

  return v.content
}

const extractPackageNameFromFiles = (files: Workspace) => {
  for (const file in files) {
    if (file.endsWith('_test.gno') || !file.endsWith('.gno')) {
      continue
    }

    if (file.includes('/')) {
      // Skip subdirs.
      continue
    }

    return parsePackageName(getWorkspaceFile(files, file))
  }
}

/**
 * Attempts to detect module name from workspace files.
 *
 * First tries gnomod.toml, then falls back to gno.mod.
 * On failure, attempts to extract package name from root files.
 *
 * @param files Workspace files.
 */
export const detectModuleName = (files: Workspace) => {
  // Try gnomod.toml first (current standard)
  if (files[GNOMOD_TOML]) {
    const modName = extractModuleName(getWorkspaceFile(files, GNOMOD_TOML))
    if (modName) {
      return modName
    }
  }

  // Fallback to legacy gno.mod
  if (files[GNOMOD_LEGACY]) {
    const modName = extractModuleName(getWorkspaceFile(files, GNOMOD_LEGACY))
    if (modName) {
      return modName
    }
  }

  const pkgName = extractPackageNameFromFiles(files)
  if (pkgName) {
    return pkgName
  }
}
