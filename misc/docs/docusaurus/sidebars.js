// @ts-check

const fs = require("fs")

/** @type {string} */
const file = fs.readFileSync("./static/sidebar.json", "utf8")

const sidebars = JSON.parse(file)

module.exports = sidebars;