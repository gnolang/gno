import { applySnapshot } from 'mobx-state-tree'

import { useStore } from '@/contexts'
import type { ExampleCategory, ExampleItem } from '@/generated/examples'

interface ExtendedExampleItem extends Omit<ExampleItem, 'isMultiFile' | 'files' | 'mainFile'> {
  isMultiFile?: boolean
  files?: Record<string, string>
  mainFile?: string
}

export function useExamplesUtilities() {
  const store = useStore()

  const reorganizeCategories = (categories: ExampleCategory[]): ExampleCategory[] => {
    // Hide archived and test examples from the menu
    const shouldShow = (item: ExampleItem) => {
      const file = item.file || ''
      if (file.includes('/archive/')) return false
      if (file.startsWith('r/tests/vm/')) return false
      return true
    }

    return categories
      .map((cat) => ({
        ...cat,
        items: cat.items.filter(shouldShow),
      }))
      .filter((cat) => cat.items.length > 0)
  }

  const filterExamples = (category: ExampleCategory, searchText: string): ExampleCategory => {
    if (!searchText) {
      return category
    }

    const searchLower = searchText.toLowerCase()

    return {
      ...category,
      items: category.items.filter((item) => {
        const titleMatch = item.title.toLowerCase().includes(searchLower)
        const descriptionMatch = item.description.toLowerCase().includes(searchLower)
        return titleMatch || descriptionMatch
      }),
    }
  }

  const generateFilename = (title: string): string => {
    return `${title.toLowerCase().replace(/\s+/g, '-')}.gno`
  }

  const handleLoadExample = (item: ExtendedExampleItem) => {
    if (!item.code) {
      console.error('Cannot load example: No code content available', item)
      return
    }

    const { code, title, isMultiFile, files, mainFile } = item
    const file = item.file

    let filename = generateFilename(title)
    let finalCode = code

    const headerRegex = /^\/\/\s*Official Example from gnolang\/gno\n\/\/\s*Source:\s*(https?:\/\/\S+)\n\n?/m
    const headerMatch = finalCode.match(headerRegex)
    if (headerMatch && headerMatch[1]) {
      const src = headerMatch[1]
      finalCode = finalCode.replace(headerRegex, '')
      finalCode = `// Gno example from: ${src}\n\n${finalCode}`
    }

    if (file) {
      const fileName = file.split('/').pop()
      if (fileName) {
        filename = fileName
      }
    }

    try {
      const workbench = store.workbench

      if (isMultiFile && files) {
        const filesObj: Record<string, { content: string; path: string }> = {}

        Object.entries(files).forEach(([fileName, content]) => {
          let finalContent = content
          if (fileName === mainFile) {
            finalContent = finalCode
          }
          filesObj[fileName] = {
            content: finalContent,
            path: fileName,
          }
        })

        const activePath = mainFile ?? Object.keys(filesObj)[0]

        applySnapshot(workbench, {
          activePath,
          files: filesObj,
          workspaceId: workbench.workspaceId,
          showFileBrowser: true,
          showTerminal: workbench.showTerminal,
          statusMessage: workbench.statusMessage,
        })
      } else {
        applySnapshot(workbench, {
          activePath: filename,
          files: {
            [filename]: {
              content: finalCode,
              path: filename,
            },
          },
          workspaceId: workbench.workspaceId,
          showFileBrowser: workbench.showFileBrowser,
          showTerminal: workbench.showTerminal,
          statusMessage: workbench.statusMessage,
        })
      }
    } catch (error) {
      console.error('Error loading example:', error)
    }
  }

  const formatItemTitle = (title: string): string => {
    let formattedTitle = title

    if (formattedTitle.startsWith('p/')) {
      formattedTitle = formattedTitle.substring(2)
    }

    if (formattedTitle.startsWith('r/')) {
      formattedTitle = formattedTitle.substring(2)
    }

    if (formattedTitle.startsWith('e/')) {
      formattedTitle = formattedTitle.substring(2)
    }

    if (formattedTitle.endsWith('.gno')) {
      formattedTitle = formattedTitle.substring(0, formattedTitle.length - 4)
    }

    return formattedTitle
  }

  return {
    reorganizeCategories,
    filterExamples,
    generateFilename,
    handleLoadExample,
    formatItemTitle,
  }
}
