/**
 * Replaces @username by [@username](/r/demo/users:username)
 * @param string rawData text to render usernames in
 * @returns string rendered text
 */
function renderUsernames(raw) {
  return raw.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/demo/users:$2)");
}
