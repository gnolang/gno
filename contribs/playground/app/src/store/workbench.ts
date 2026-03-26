import { compress, createSession, filterValidFiles, getSession, isReservedFileName } from '@gnostudio/core'
import { cloneGitRepository, generateRandom } from '@gnostudio/pkg'

import { reaction } from 'mobx'
import {
  addDisposer,
  applySnapshot,
  destroy,
  detach,
  flow,
  getRoot,
  getSnapshot,
  types,
  type Instance,
} from 'mobx-state-tree'

import {
  DuplicateFileNameException,
  InvalidExtensionException,
  InvalidNameException,
  ReservedFileNameException,
} from '@/lib'

import { type RootStore } from '.'
import { MessageType, StatusMessageModel, type StatusMessage } from './types/status-message'

export const File = types
  .model({
    path: types.identifier,
    content: types.optional(types.string, ''),
  })
  .volatile(() => ({
    hasUnsavedChanges: false,
  }))
  .actions((self) => ({
    setContent(content: string) {
      self.content = content
    },

    markSaved() {
      self.hasUnsavedChanges = false
    },

    afterCreate() {
      const dispose = reaction(
        () => self.content,
        () => {
          self.hasUnsavedChanges = true
        },
      )

      addDisposer(self, dispose)
    },
  }))

export type FileType = Instance<typeof File>

export const Workbench = types
  .model({
    activePath: types.optional(types.string, '/'),
    files: types.map(File),
    workspaceId: types.optional(types.string, Date.now().toString()),
    showFileBrowser: types.optional(types.boolean, true),
    showTerminal: types.optional(types.boolean, false),
    statusMessage: types.maybeNull(StatusMessageModel),
  })
  .volatile(() => ({
    renamingPath: undefined as string | undefined,
    pendingFileName: false as string | boolean,
    sharedUri: undefined as string | undefined,
    revisionHash: '', // Used to detect unsaved changes
  }))
  .views((self) => ({
    get tabs() {
      return Array.from(self.files.values())
    },

    get activeFile() {
      let path = self.activePath

      // Fallback to defaults if file doesn't exist or bad query parameter.
      if (!self.files.has(path)) {
        path = self.files.keys().next().value
      }
      return self.files.get(path)
    },

    get isRenaming() {
      return self.renamingPath !== undefined
    },

    get isCreatingFile() {
      return self.pendingFileName !== false && self.renamingPath === undefined
    },

    get shareUrl() {
      if (!self.sharedUri) return undefined

      const { origin } = new URL(import.meta.url)
      return new URL(`/p/${self.sharedUri}`, origin).href
    },

    getGitUrl(provider: string, owner?: string, repo?: string) {
      if (!owner || !repo) return undefined

      switch (provider) {
        case 'github':
          return `https://github.com/${String(owner)}/${String(repo)}`
      }
    },
  }))
  .actions((self) => ({
    validatePendingFileName() {
      if (self.files.has(self.pendingFileName as string)) {
        throw new DuplicateFileNameException()
      }

      const validFileName = /^[a-z0-9-_.()]+$/i
      const validExtension = /(gnomod\.toml|\.gno)$/i

      if (isReservedFileName(self.pendingFileName as string)) {
        throw new ReservedFileNameException()
      }

      if (!validFileName.test(self.pendingFileName as string)) {
        throw new InvalidNameException()
      }

      if (!validExtension.test(self.pendingFileName as string)) {
        throw new InvalidExtensionException()
      }

      return true
    },
    setExtension() {
      const hasExt = (self.pendingFileName as string).includes('.')
      if (!hasExt) {
        self.pendingFileName = (self.pendingFileName as string) + '.gno'
      }
    },
    setStatusMessage(msg: StatusMessage | null) {
      self.statusMessage = msg
    },
  }))
  .actions((self) => ({
    setActivePath(path: string) {
      self.activePath = path
    },

    setFiles(files: Record<string, File>) {
      self.files.replace(files)
    },

    updateRevisionHash() {
      self.revisionHash = generateRandom()
    },

    updateFileContent(path: string, content: string) {
      self.files.get(path)?.setContent(content)
    },

    dropFiles: flow(function* (files: File[]) {
      for (const file of files) {
        const content = yield file.text()
        self.files.put({ path: file.name, content })
      }

      self.activePath = files[0].name
    }),

    reorderFile(path: string, index: number) {
      const files = getSnapshot(self.files)
      const paths = Object.keys(files)
      const tabIndex = paths.indexOf(path)

      // Remove the file from the array and insert it at the new index
      const [file] = paths.splice(tabIndex, 1)
      paths.splice(index, 0, file)

      const newFiles = paths.reduce((acc, key) => ({ ...acc, [key]: files[key] }), {})
      self.files.replace(newFiles)
    },

    setPendingFileName(name: string) {
      self.pendingFileName = name
    },

    startAddFile() {
      self.renamingPath = undefined
      self.pendingFileName = true
    },

    startRenameFile(path: string) {
      self.renamingPath = path
      self.pendingFileName = path
    },

    renameFile(oldPath: string, newPath: string) {
      if (!self.validatePendingFileName()) return

      const files = getSnapshot(self.files)
      const paths = Object.keys(files)

      const newFiles = paths.reduce((acc, key) => {
        const path = key === oldPath ? newPath : key
        const file = {
          ...files[key],
          path,
        }

        return {
          ...acc,
          [path]: file,
        }
      }, {})

      detach(self.files.get(oldPath))
      self.files.replace(newFiles)

      // Update the active path if it's the renamed file
      if (self.activePath === oldPath) {
        self.activePath = newPath
      }

      self.pendingFileName = false
      self.renamingPath = undefined
    },

    addFile() {
      // Handle the case where the user cancels the file creation
      if (self.pendingFileName === true) return

      self.setExtension()

      if (!self.validatePendingFileName()) return

      self.files.put({
        path: self.pendingFileName as string,
        content: '',
      })

      self.activePath = self.pendingFileName as string
      self.pendingFileName = false
    },

    cancelNameFile() {
      self.pendingFileName = false
      self.renamingPath = undefined
    },

    deleteFile(path: string) {
      if (self.files.size === 1) {
        alert('You cannot delete the last file.')
        return
      }

      const result = confirm(`Are you sure you want to delete\n${String(path)}?`)
      if (!result) return

      const file = self.files.get(path)
      const prevTabIndex = self.tabs.findIndex((tab) => tab.path === path)

      self.files.delete(path)

      // If the file is active, select the next tab
      if (self.activePath === path) {
        const tabIndex = prevTabIndex - 1
        const nextTab = self.tabs[tabIndex + 1] || self.tabs[tabIndex - 1] || self.tabs[0]
        self.activePath = nextTab.path
      }

      destroy(file)
    },

    clearSharedUri() {
      self.sharedUri = undefined
    },
  }))
  .actions((self) => ({
    saveToCloud: flow(function* saveToCloud() {
      // The sharedUri will change when the files change
      if (self.sharedUri) return self.sharedUri

      const response = yield createSession({
        description: '',
        files: Object.values(getSnapshot(self.files)),
      })

      // Response URI contains a hash with GCS session ID
      self.sharedUri = response.uri

      return response.uri
    }),

    loadFromCloud: flow(function* loadFromCloud(uri: string) {
      const response = yield getSession(uri)

      self.files.clear()

      for (const file of filterValidFiles(response.files)) {
        self.files.put(file)
      }

      self.workspaceId = uri
      self.sharedUri = uri
      self.activePath = response.files[0]?.path
    }),

    loadFromProject: flow(function* loadFromProject(projectId: string) {
      const root = getRoot(self)
      const project = (root as RootStore).projects.loadById(projectId)

      if (!project) {
        throw new Error('Invalid project ID')
      }

      self.files.clear()
      applySnapshot(self.files, project.files)

      self.workspaceId = projectId
      self.activePath = self.tabs[0]?.path
    }),

    loadFromGit: flow(function* loadFromGit(provider: string, owner?: string, repo?: string) {
      const gitUrl = self.getGitUrl(provider, owner, repo)

      if (!gitUrl) {
        throw new Error('Invalid git url')
      }

      try {
        self.setStatusMessage({ type: MessageType.Progress, tag: 'Import', text: `Importing from ${provider}...` })

        const response = yield cloneGitRepository(gitUrl)
        self.files.clear()

        for (const file of filterValidFiles(response)) {
          const path = file.path.replace('./', '')
          self.files.set(path, { path, content: file.content })
        }

        self.activePath = self.tabs[0]?.path
        self.workspaceId = gitUrl

        const importedCount = self.files.size
        self.setStatusMessage({
          type: MessageType.Info,
          tag: 'Import',
          text: `Imported ${importedCount} file${importedCount === 1 ? '' : 's'} from ${provider}`,
        })

        return response
      } catch (error) {
        self.setStatusMessage({ type: MessageType.Error, tag: 'Import', text: `Failed to import from ${provider}` })
        throw error
      } finally {
        // Auto-clear status after a short delay
        yield new Promise((resolve) => setTimeout(resolve, 3000))
        self.setStatusMessage(null)
      }
    }),

    // Import a single file from GitHub
    loadSingleFileFromGithub: flow(function* loadSingleFileFromGithub(
      owner: string,
      repo: string,
      path: string,
      branch = 'main',
    ) {
      try {
        self.setStatusMessage({ type: MessageType.Progress, tag: 'Import', text: 'Fetching file from GitHub...' })
        const rawUrl = `https://raw.githubusercontent.com/${owner}/${repo}/refs/heads/${branch}/${path}`
        const response = yield fetch(rawUrl)

        if (!response.ok) {
          throw new Error(`Failed to fetch file '${path}' (status ${response.status}): ${response.statusText}`)
        }

        const content = yield response.text()
        self.files.clear()

        const filename = path.split('/').pop() ?? path
        self.files.put({ path: filename, content })
        self.activePath = filename
        self.workspaceId = `github:${owner}/${repo}/${branch}/${path}`
        self.setStatusMessage({ type: MessageType.Info, tag: 'Import', text: `Imported ${filename} from GitHub` })

        return { path: filename, content }
      } catch (error) {
        console.error('Error importing single file from GitHub:', error)
        self.setStatusMessage({ type: MessageType.Error, tag: 'Import', text: 'Failed to import file from GitHub' })
        throw error
      } finally {
        // Auto-clear status after a short delay
        yield new Promise((resolve) => setTimeout(resolve, 3000))
        self.setStatusMessage(null)
      }
    }),

    loadFromSerializedHash(hash: string) {
      let files = JSON.parse(compress.atou(hash))
      self.files.clear()

      // Cast to array if not already
      files = Array.isArray(files) ? files : [files]

      for (const file of filterValidFiles(files)) {
        self.files.set(file.path, {
          path: file.path,
          content: file.content,
        })
      }

      self.activePath = files[0]?.path
      self.workspaceId = hash
    },

    afterCreate() {
      const dispose = reaction(
        () => getSnapshot(self.files),
        () => {
          self.updateRevisionHash()
          self.clearSharedUri()
        },
        {
          delay: 500,
        },
      )

      addDisposer(self, dispose)
    },
  }))
