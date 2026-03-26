import type { Plugin } from 'vite'

export function minifyPlugin(skipObfuscation?: boolean): Plugin

export function checkWorkspaceDependencies(deps: string[]): Promise<boolean>
