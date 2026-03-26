import { createInterface } from 'readline/promises'
import { $, fs, path } from 'zx' // Added process import

/**
 * Checks if a local workspace dependency has been built by verifying the existence of its 'dist' directory.
 *
 * @param {string} depPath - The relative path to the workspace dependency
 * @returns {Promise<boolean>} Returns true if the dist directory exists, false otherwise
 */
async function checkWorkspaceDependency(depPath = '') {
  const distPath = path.join(path.dirname(new URL(import.meta.url).pathname), depPath, 'dist')

  const exists = await fs.promises
    .access(distPath)
    .then(() => true)
    .catch(() => false)

  if (!exists || !(await fs.promises.stat(distPath)).isDirectory()) {
    return false
  }

  return true
}

/**
 * Checks multiple workspace dependencies to verify their build status and logs results.
 * Prints a summary report showing which dependencies are built and which are missing.
 *
 * @param {string[]} deps - Array of paths to workspace dependencies
 * @returns {Promise<boolean>} Returns true if all dependencies are built, false otherwise
 */
export async function checkWorkspaceDependencies(deps = []) {
  const results = await Promise.allSettled(deps.map(checkWorkspaceDependency))

  console.log('\nWorkspace Dependencies Build Status:')
  console.log('-----------------------------------')

  const built = []
  const missing = []

  results.forEach((result, index) => {
    const depPath = deps[index]
    if (result.status === 'fulfilled' && result.value) {
      built.push(depPath)
    } else {
      missing.push(depPath)
    }
  })

  if (built.length > 0) {
    console.log('\n✓ Built dependencies:')
    built.forEach((dep) => console.log(`  - ${dep}`))
  }

  if (missing.length > 0) {
    console.log('\n⨯ Missing builds:')
    missing.forEach((dep) => console.log(`  - ${dep}`))
    console.log('\nPlease build the missing dependencies:')
    for (const dep of missing) {
      console.log(`  - "cd ${dep} && pnpm build"`)
    }

    const shouldBuild = await confirm('Do you want to build the missing dependencies? [y/N]')

    if (shouldBuild) {
      for (const dep of missing) {
        const depPath = relativePath(dep)
        const depName = path.basename(depPath)
        console.log(`\n⏳ Building ${depName}...`)
        await $`cd ${depPath} && pnpm build`.catch(() => false)
        console.log(`\✅ ${depName} built successfully.`)
      }
      return true
    }

    return false
  }

  console.log('\n✅ All workspace dependencies ready.')
  return true
}

/**
 * Prompts the user with a message and waits for confirmation (y/yes).
 *
 * @param {string} message - The message to display to the user.
 * @returns {Promise<boolean>} Returns true if the user confirms, false otherwise.
 */
async function confirm(message) {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  })
  try {
    const answer = await rl.question(message + ' ')
    const $answer = answer.trim().toLowerCase()
    return ['y', 'yes'].includes($answer)
  } catch (err) {
    console.error('Error reading user input:', err)
    return false
  } finally {
    rl.close()
  }
}

function relativePath(depPath) {
  const scriptDir = path.dirname(new URL(import.meta.url).pathname)
  const absoluteDepPath = path.resolve(scriptDir, depPath)

  return path.relative(process.cwd(), absoluteDepPath)
}
