import { go } from '@codemirror/lang-go'
import { Compartment, type Extension } from '@codemirror/state'

import { goMod } from './modfile'

export enum Syntax {
  Gno,
  ModFile,
  Unknown,
}

export const defaultSyntax = Syntax.Gno

const getFileExtension = (fName?: string) => {
  const extPos = fName?.lastIndexOf('.') ?? -1
  return extPos !== -1 && fName?.toLowerCase()?.slice(extPos)
}

/**
 * Detects language for syntax highlight by extension in a file name.
 */
export const syntaxFromFileName = (fName?: string): Syntax => {
  switch (getFileExtension(fName)) {
    case '.mod':
      return Syntax.ModFile
    case '.gno':
      return Syntax.Gno
    default:
      return Syntax.Unknown
  }
}

const getSyntaxExtension = (lang: Syntax): Extension => {
  switch (lang) {
    case Syntax.Gno:
      return go()
    case Syntax.ModFile:
      return goMod()
    default:
      return []
  }
}

export const syntaxCompartment = new Compartment()

/**
 * Returns a new syntax extension compartment with initial syntax highlighter based on a file name.
 */
export const newSyntaxCompartment = (fileName?: string) =>
  syntaxCompartment.of(getSyntaxExtension(syntaxFromFileName(fileName)))

/**
 * Returns a new EditorState effect to replace syntax highlight extension.
 *
 * Use `syntaxFromFileName` to get language.
 */
export const updateSyntaxEffect = (lang: Syntax) => syntaxCompartment.reconfigure(getSyntaxExtension(lang))
