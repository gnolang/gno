import * as BrowserFS from 'browserfs'
import type InMemoryFileSystem from 'browserfs/dist/node/backend/InMemory'
import type MountableFileSystem from 'browserfs/dist/node/backend/MountableFileSystem'
import type OverlayFSFileSystem from 'browserfs/dist/node/backend/OverlayFS'
import type ZipFSFileSystem from 'browserfs/dist/node/backend/ZipFS'
import type { FileSystem } from 'browserfs/dist/node/core/file_system'
import promisfy from 'pify'

import { WorkerFS } from './workerfs'

type InMemory = (options?: any) => Promise<InMemoryFileSystem>
type Mountable = (options?: any) => Promise<MountableFileSystem>
type OverlayFS = (options?: any) => Promise<OverlayFSFileSystem>
type ZipFS = (options: Parameters<typeof BrowserFS.FileSystem.ZipFS.Create>[0]) => Promise<ZipFSFileSystem>

export const BackendInmemory: InMemory = promisfy(BrowserFS.FileSystem.InMemory.Create)
export const BackendMountableFileSystem: Mountable = promisfy(BrowserFS.FileSystem.MountableFileSystem.Create)
export const BackendOverlayFS: OverlayFS = promisfy(BrowserFS.FileSystem.OverlayFS.Create)
export const BackendZipFS: ZipFS = promisfy(BrowserFS.FileSystem.ZipFS.Create)
export const BackendWorkerFS = promisfy(WorkerFS.Create)

/**
 * Maps incoming requests from WorkerFS to a filesystem.
 *
 * This function should be used on filesystem host side.
 *
 * @param fs BFS file system instance
 * @param listener Message port or Worker instance.
 */
export const attachWorkerFSListener = (fs: FileSystem, listener: MessagePort | Worker) =>
  WorkerFS.attachRemoteListener(fs, listener)
