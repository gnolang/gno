import { types } from 'mobx-state-tree'

export const DEFAULT_RUN_EXPRESSION = 'main()'

/**
 * EvalMode descibes current project evaluation mode
 */
export enum EvalMode {
  /**
   * None means that program is not running.
   */
  None = 'none',

  /**
   * Repl indicates that program is running inside REPL.
   */
  Repl = 'repl',

  /**
   * Test indicates that program is running inside unit test runner.
   */
  Test = 'test',

  /**
   * Test indicates that "gno run" should be called with current stored expression.
   */
  Run = 'run',
}

export const EvalState = types
  .model({
    isRunPromptOpen: types.optional(types.boolean, false),
    evalExpression: types.optional(types.string, ''),
    evalMode: types.optional(types.enumeration(Object.values(EvalMode)), EvalMode.None),
  })
  .views((self) => ({
    get isRepl() {
      return self.evalMode === EvalMode.Repl
    },
  }))
  .actions((self) => ({
    setEvalMode(newMode: EvalMode) {
      self.evalMode = newMode
    },
    resetEvalMode() {
      self.evalMode = EvalMode.None
    },
    setEvalExpression(expr: string) {
      self.evalMode = EvalMode.Run
      self.evalExpression = expr
    },
    setRunPromptOpen(isVisible: boolean) {
      self.isRunPromptOpen = isVisible
    },
  }))
