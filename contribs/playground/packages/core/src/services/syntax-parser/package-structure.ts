import { go } from '@codemirror/lang-go'
import { type SyntaxNode } from '@lezer/common'

const goParser = go().language.parser

export interface FunctionSymbol {
  name: string
  return?: string
  private: boolean
  signature: string
  params: Array<{ name?: string; type?: string }>
  pos: {
    from: number
    to: number
  }
}

/**
 * Extracts the document structure from a Gno code.
 * This is used to provide code intelligence features such as autocomplete, jump to definition, etc.
 *
 * @param code - The Go code to analyze
 * @returns An object containing the package, imports, and functions
 */
export function derivePackageStructure(code: string) {
  const tree = goParser.parse(code)
  const cursor = tree.topNode.cursor()

  const packageNode = cursor.node.getChild('PackageClause')
  const importNode = cursor.node.getChild('ImportDecl')
  const functionNodes = cursor.node.getChildren('FunctionDecl')

  return {
    package: derivePackageSymbol(packageNode, code),
    imports: deriveImportSymbols(importNode, code),
    functions: functionNodes.map((node) => deriveFunctionSymbol(node, code)).filter(Boolean) as FunctionSymbol[],
  }
}

function derivePackageSymbol(node: SyntaxNode | null, code: string): string | undefined {
  if (!node) return undefined
  return getNodeText(node.getChild('DefName'), code)
}

function deriveImportSymbols(node: SyntaxNode | null, code: string): string[] {
  if (!node) return []
  const c = node.cursor()
  const imports: Array<string | undefined> = []

  if (!c.firstChild()) return []

  while (c.nextSibling()) {
    switch (c.name) {
      case 'SpecList': {
        const specNodes = c.node.getChildren('ImportSpec')
        imports.push(...specNodes.map((spec) => getNodeText(spec, code)))
        break
      }
      case 'ImportSpec': {
        imports.push(getNodeText(c.node, code))
        break
      }
    }
  }

  return imports.filter(Boolean).map((v) => (v as string).replace(/"/g, ''))
}

/**
 * Returns a function symbol object derived from the function declaration node.
 */
function deriveFunctionSymbol(node: SyntaxNode, code: string): FunctionSymbol | null {
  const result: FunctionSymbol = {
    name: '',
    return: undefined,
    private: false,
    signature: '',
    params: [],
    pos: {
      from: node.from,
      to: node.to,
    },
  }

  const c = node.cursor()
  if (!c.firstChild()) return null

  let prevNode = c.node
  let parametersText = ''

  while (c.nextSibling()) {
    switch (c.name) {
      case 'DefName':
        result.name = getNodeText(c.node, code) as string
        break
      case 'TypeName':
        result.return = getNodeText(c.node, code) as string
        break
      case 'Parameters': {
        const isReturnParameters = prevNode.name === 'Parameters'

        if (isReturnParameters) {
          result.return = getNodeText(c.node, code) as string
          break
        }

        parametersText = getNodeText(c.node, code) as string
        const parameters = c.node.getChildren('Parameter') ?? []

        for (const param of parameters) {
          const paramText = getNodeText(param, code)?.split(' ') ?? []

          result.params.push({
            name: paramText[0],
            type: paramText[1],
          })
        }

        break
      }
      case 'PointerType': {
        result.return = getNodeText(c.node, code) as string
      }
    }
    prevNode = c.node
  }

  // If the function name starts with a lowercase letter, it's private
  result.private = result.name.charAt(0) === result.name.charAt(0).toLowerCase()
  result.signature = [`${result.name}${parametersText}`, result.return].filter(Boolean).join(' ')
  return result
}

/**
 * Helper function to get the text of a node
 * It uses function overloading to handle null nodes
 */
function getNodeText(node: SyntaxNode | null, code: string) {
  if (!node) return undefined
  return code.slice(node.from, node.to)
}
