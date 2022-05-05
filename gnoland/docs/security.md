# gno.land Website

The gno.land website has 3 main dependencies:

1. [UmbrellaJs](https://umbrellajs.com/) for DOM operations
2. [MarkedJs](https://marked.js.org/) for Markdown to html compilation
3. [HighlightJs](https://highlightjs.org/) for golang syntax highlighting
4. [DOMPurify](https://github.com/cure53/DOMPurify) to sanitize html (and avoid xss)

Some security considerations:
| | Umbrella Js | Marked Js | HighlightJs | DOMPurify |
|---|---|---|---|---|
| dependencies | 0 | 0 | 0 | 0 |
| sanitize content | | [no](https://marked.js.org/#usage) | [throws an error](https://github.com/highlightjs/highlight.js/blob/7addd66c19036eccd7c602af61f1ed84d215c77d/src/highlight.js#L741) | [yes](https://github.com/cure53/DOMPurify#readme) |

Best Practices:

- **When using MarkedJs**: Always run the output of the marked compiler inside `DOMPurify.sanitize` before inserting it in the dom with `.innerHtml = `.
- **When using DOMPurify**: Preferably use `{ USE_PROFILES: { html: true } }` option to allow html only. Content passed in the sanitizer must not be modified afterwards, and must directly be inserted in the DOM with innerHtml. Do not call `DOMPurify.sanitize` with the output of a previous `DOMPurify.sanitize` to avoid any mutation XSS risks.
- **When using HighlightJs**: always configure it before with `hljs.configure({throwUnescapedHTML: true})` to throw before inserting html in the page if any unexpected html children are detected. The check is done [here](https://github.com/highlightjs/highlight.js/blob/7addd66c19036eccd7c602af61f1ed84d215c77d/src/highlight.js#L741).
