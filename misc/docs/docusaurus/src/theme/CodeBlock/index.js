import React from "react";
import CodeBlock from "@theme-original/CodeBlock";
import { useColorMode } from "@docusaurus/theme-common";
import {
  parseContentToFiles,
  parseMetaAttributes,
} from "../../remark/tabbed-code-block";

export default function CodeBlockWrapper(props) {
  const { colorMode } = useColorMode();
  const meta = parseMetaAttributes(props.metastring ?? "");
  const { path, run_expr } = meta;

  if (props.className === "language-go" && "gno" in meta) {
    const files = parseContentToFiles(props.children, path);

    return (
      <gs-playground
        files={JSON.stringify(files)}
        theme={colorMode}
        run-expression={run_expr}
        menu-always-open
      />
    );
  }

  return (
    <>
      <CodeBlock {...props} />
    </>
  );
}
