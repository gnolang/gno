import React, { useEffect } from "react";
// import "./style.css";
import "meilisearch-docsearch/css";

import useDocusaurusContext from '@docusaurus/useDocusaurusContext';


export default function Component() {
  const { siteConfig } = useDocusaurusContext();

  const {
    meilisearchURL,
    meilisearchApiKey,
    meilisearchIndexUid,
  } = siteConfig.customFields

  useEffect(() => {
    const lang = document.querySelector("html").lang || "en";

    const docsearch = require("meilisearch-docsearch").default;
    const destroy = docsearch({
      host: meilisearchURL,
      apiKey: meilisearchApiKey,
      indexUid: meilisearchIndexUid,
      container: "#docsearch",
      //   searchParams: {filter: [`lang = ${lang}`]},
    });

    return () => destroy();
  }, []);

  return <div id="docsearch" />;
}
