interface FormatIssue {
  line: number
  column: number
  message: string
}

/**
 * Parses gofmt's error output into a list of errors.
 * @param output stdout output
 */
export const parseErrorOutput = (output: string): FormatIssue[] => {
  if (!output) {
    return []
  }

  return output.split('\n').reduce<FormatIssue[]>((issues, line) => {
    const [row, column, message] = line.replace(/^<standard input>:/, '').split(':', 3)
    if (!message) {
      return issues
    }

    const issue: FormatIssue = {
      line: Number.parseInt(row),
      column: Number.parseInt(column),
      message: message.trim(),
    }

    if (issue.line && issue.column && issue.message) {
      issues.push(issue)
    }

    return issues
  }, [])
}
