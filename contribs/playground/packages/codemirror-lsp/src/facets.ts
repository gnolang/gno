import { Facet } from '@codemirror/state'

import type { LanguageServerClient } from './client/client'
import { useLast } from './utils'

export const client = Facet.define<LanguageServerClient, LanguageServerClient>({
  combine: useLast,
})

export const documentUri = Facet.define<string, string>({ combine: useLast })
export const languageId = Facet.define<string, string>({ combine: useLast })
