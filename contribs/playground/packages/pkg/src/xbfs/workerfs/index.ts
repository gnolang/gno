/**
 * WorkerFS implementation copy from BrowserFS package.
 *
 * Forked as for some unknown reason, Vite bundles 2 copies of a same module within a single JS chunk
 * causing breakage of WorkerFS driver.
 *
 * **Context:**
 *
 * During FS call WorkerFS checks each argument type to marshal it properly.
 * Check is done using 'instanceof' operator. If check fails, it returns 'Invalid argument' string value.
 *
 * @see `_argLocal2Remote` in `workerfs.ts` as line 566.
 *
 * **Root cause:**
 *
 * Inside BFS, modules are resolved from `/node_modules/.pnpm/browserfs@2.0.0/node_modules/browserfs/src/core/FILE_NAME.ts`,
 * but in our code, modules are resolved from `/node_modules/.pnpm/browserfs@2.0.0/src/core/FILE_NAME.ts` (which doesn't exist).
 *
 * As Vite provides a different copy of BFS modules to our code and BFS inself (internal import),
 * this breaks 'instanceof' check and makes WorkerFS unusable.
 *
 * I tried everything:
 * - pnpm's dependency flattening.
 * - `node-linker=hoisted`.
 * - `resolve.dedupe`.
 * - hardcoding module path using `resolve.alias`.
 * - and many other options.
 *
 * But nothing helps.
 *
 * **Solution:**
 *
 * The only option atm is to just "fork" WorkerFS with our own patches to get the job done
 * and hope that someone in future will be able to solve this.
 *
 * P.S - some certain eslint rules are disabled to avoid edit or break of original BFS code.
 */

export { WorkerFS } from './workerfs'
