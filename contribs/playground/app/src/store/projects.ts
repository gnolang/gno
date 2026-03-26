/* eslint-disable @typescript-eslint/no-unnecessary-type-assertion */
import { generateRandom } from '@gnostudio/pkg'

import { getRoot, getSnapshot, types, type Instance } from 'mobx-state-tree'
import { persist } from 'mst-persist'

import { type RootStore } from '.'
import { File } from './workbench'

export const Project = types
  .model({
    id: types.identifier,
    files: types.frozen(types.map(File)),
    title: types.maybe(types.string),
    isDraft: types.optional(types.boolean, true),
    timestamp: types.optional(types.number, Date.now()),
  })
  .volatile(() => ({
    revisionHash: '',
  }))

export type ProjectType = Instance<typeof Project>

export const Projects = types
  .model({
    projects: types.map(Project),
  })
  .volatile(() => ({
    activeId: '' as string | undefined,
  }))
  .views((self) => ({
    get hasActive() {
      return !!self.activeId
    },
    get active() {
      if (!self.activeId) return undefined
      return self.projects.get(self.activeId)
    },
    get list() {
      return Array.from(self.projects.values()).sort((a, b) => b.timestamp - a.timestamp)
    },
  }))
  .views((self) => ({
    get hasUnsavedChanges() {
      const root = getRoot(self)
      const workbenchHash = (root as any).workbench.revisionHash

      // Ignore the initial state
      if (workbenchHash === '') return false
      if (!self.active) return true

      return self.active.revisionHash !== workbenchHash
    },
  }))
  .actions((self) => ({
    loadById(id: string) {
      const project = self.projects.get(id)
      if (!project) return

      self.activeId = id
      return project
    },

    saveDraft({ id, title }: { id?: string; title?: string } = {}) {
      const projectId = id ?? generateRandom()
      const workbench = (getRoot(self) as RootStore).workbench

      const files = getSnapshot(workbench.files)

      const project = self.projects.put({ id: projectId, title, files, isDraft: true })
      project.revisionHash = workbench.revisionHash
      project.timestamp = Date.now()

      workbench.files.forEach((file) => file.markSaved())

      self.activeId = projectId
      return project
    },

    saveActiveProject() {
      if (!self.active) return
      const workbench = (getRoot(self) as RootStore).workbench

      self.active.files = getSnapshot(workbench.files)
      self.active.revisionHash = workbench.revisionHash
      self.active.timestamp = Date.now()

      workbench.files.forEach((file) => file.markSaved())
    },

    delete(id: string) {
      const confirm = window.confirm('Are you sure you want to delete this project?')
      if (!confirm) return

      self.projects.delete(id)

      if (self.activeId === id) {
        self.activeId = undefined
      }
    },
  }))
  .actions((self) => ({
    saveDeployedProject({ id }: { id: string }) {
      const oldActiveId = self.activeId

      const draft = self.saveDraft({ id, title: id })
      draft.isDraft = false

      if (oldActiveId) {
        self.projects.delete(oldActiveId)
      }
    },

    save() {
      if (self.active) {
        return self.saveActiveProject()
      }

      return self.saveDraft()
    },

    update(id: string, { title }: { title?: string } = {}) {
      const project = self.projects.get(id)
      if (!project) return

      project.title = title
      project.timestamp = Date.now()
    },

    afterCreate() {
      // Prevent accidental navigation away from the page
      window.addEventListener('beforeunload', (e) => {
        if (import.meta.env.VITE_DISABLE_BEFORE_UNLOAD) {
          return
        }

        if (self.hasUnsavedChanges) {
          e.preventDefault()
          e.returnValue = true
        }
      })

      persist('projects', self)
        .then(() => {})
        .catch(console.error)
    },
  }))
