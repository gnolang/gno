/**
 * List of known editor context menu actions.
 */
export enum ContextMenuAction {
  /**
   * Open a Run dialog with expression prompt.
   */
  OpenRunPrompt = 'gnovm.expr.prompt',

  /**
   * Run previously called expression.
   */
  RunLastAction = 'gnovm.expr.replay',
}
