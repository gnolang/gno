/**
 * MapLike is interface similar to builtin's Map but allows other implementations.
 *
 * Used to support map-like structs such as mst's IMSTMap.
 */
interface MapLike<K, V> {
  has: (k: K) => boolean
  get: (k: K) => V | undefined
  set: (k: K, v: V) => void
  keys: () => IterableIterator<K>
  delete: (k: K) => void
}

export interface FileNode {
  path: string
  content: string
}

export type FilesMap = MapLike<string, FileNode>

export interface DirEntry {
  children: Set<string>
}
