import { visit } from "unist-util-visit";

/**
 * A transformer function for processing Gno code blocks from Markdown.
 * This detects Gno-specific code blocks, manages dependencies between them,
 * and modifies the content of dependent blocks by including the content of their dependencies.
 *
 * Example:
 * ```go gno path=main.gno
 * package main
 *
 * func Render(_ string) string {
 *  return "Hello World"
 * }
 * ```
 *
 * ```go gno path=main_test.gno depends_on=main.gno
 * package main
 *
 * import "testing"
 * // ...
 * ```
 */
export default function tabbedCodeBlock() {
  function isGnoCodeBlock(node) {
    const { type, lang } = node ?? {};
    const meta = parseMetaAttributes(node.meta);
    return type === "code" && lang === "go" && "gno" in meta;
  }

  const nodes = {};

  return (tree) => {
    visit(tree, isGnoCodeBlock, (node) => {
      const metaAttrs = parseMetaAttributes(node.meta);
      const { path, depends_on } = metaAttrs;
      if (!path) return;

      nodes[path] = node;

      const fileDependency = nodes[depends_on];
      if (!fileDependency) return;

      // Construct the combined content with dependency information
      // using the magic comment `//=== "filename"` to declare multiple
      // files in the same block.
      let output = `//=== "${depends_on}"\n`;
      output += fileDependency.value;
      output += `\n//=== "${path}"\n`;
      output += node.value;

      node.value = output;
    });
  };
}

/**
 * Parses meta attributes from a string and returns them as a key-value object.
 */
export function parseMetaAttributes(meta) {
  if (!meta) return {};

  return meta.split(" ").reduce((acc, attr) => {
    const [key, value] = attr.split("=");
    return { ...acc, [key]: value };
  }, {});
}

/**
 * Parses a code block containing multiple files and extracts them into an object of files.
 * The files must be separated using the magic comment `//=== "filename"`.
 *
 * Example:
 * ```go
 * //=== "main.gno"
 * package main
 * ...
 * //=== "main_test.gno"
 * ...
 * //=== "utils.gno"
 * ```
 */
export function parseContentToFiles(value, defaultPath = "main.gno") {
  const lines = value.split("\n");
  const files = {};

  let currentTab = defaultPath;

  for (const line of lines) {
    const match = line.match(/^\/\/=== "(.*)"/);
    if (match) {
      currentTab = match[1];
      files[currentTab] = "";
    } else {
      files[currentTab] = `${(files[currentTab] ?? "") + line}\n`;
    }
  }

  for (const filename in files) {
    files[filename] = files[filename].trimEnd();
  }

  return files;
}
