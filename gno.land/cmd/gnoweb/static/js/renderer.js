/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}

function parseContent(source) {
    marked.setOptions({ gfm: true });
    const doc = new DOMParser().parseFromString(source, "text/html");
    const contents = doc.documentElement.textContent;
    return marked.parse(contents);
}
