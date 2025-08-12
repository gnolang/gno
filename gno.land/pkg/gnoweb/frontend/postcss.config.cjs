const postcssImport = require("postcss-import");
const postcssPresetEnv = require("postcss-preset-env");
const cssnano = require("cssnano");
const { purgeCSSPlugin } = require("@fullhuman/postcss-purgecss");

const isProd = process.env.NODE_ENV === "production";

const plugins = [
	postcssImport(), // Must be first to process @import
	postcssPresetEnv({
		stage: 1,
		features: {
			"nesting-rules": true,
			"custom-media-queries": true,
			"custom-properties": false,
			"logical-properties-and-values": true,
			"has-pseudo-class": true,
		},
		autoprefixer: { grid: "autoplace" },
	}),
];

if (isProd) {
	plugins.push(
		purgeCSSPlugin({
			content: [
				"../**/*.html", // gnoweb HTML templates
				"../**/*.go", // gnoweb Go code
				"./js/**/*.ts", // frontend JavaScript
			],
			// keep dynamic classes
			safelist: [
				/^is-/, // states
				/^has-/, // states
				/-active$/, // states
				/-open$/, // states
				"hidden", // utils
				"dev-mode", // utils
			],
			defaultExtractor: (content) => content.match(/[\w-:/%.]+(?<!:)/g) || [],
		}),
		cssnano({ preset: "default" }),
	);
}

module.exports = { plugins };
