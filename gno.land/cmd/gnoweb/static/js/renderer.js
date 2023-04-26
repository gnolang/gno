/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}

const components = [
    {name: 'jumbotron', toRender: content => `<div class="comp-jumbotron">${content}</div>`},
]

const extensionBuilder = (comp) => {
    const {name, toRender} = comp
    const startReg = RegExp(`:::${name}`);
    const tokenizerReg = RegExp(`^:::${name}\n([\\s\\S]*?)\n:::`);
    return {
        name: name,
        level: 'block',
        start(src) { return src.match(startReg)?.index; },
        tokenizer(src, tokens) {
          const match = tokenizerReg.exec(src);
          if (match) {
            const token = {
              type: name,
              raw: match[0],
              text: match[1].trim(),
              tokens: []
            };
            this.lexer.blockTokens(token.text, token.tokens);
            return token;
          }
        },
        renderer(token) {
          return toRender(this.parser.parse(token.tokens));
        }
    };
}

function parseContent(source) {
    marked.setOptions({ gfm: true });
    components.forEach(comp => marked.use({ extensions: [extensionBuilder(comp)] }))
    const doc = new DOMParser().parseFromString(source, "text/html");
    const contents = doc.documentElement.textContent;
    return marked.parse(contents);
}