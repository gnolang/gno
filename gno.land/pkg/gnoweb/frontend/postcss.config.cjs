// frontend/postcss.config.cjs
const path = require("node:path");
const postcssImport = require("postcss-import");
const postcssPresetEnv = require("postcss-preset-env");
const cssnano = require("cssnano");
const purge = require("@fullhuman/postcss-purgecss");
const purgecss = purge.default || purge;

const here = (...p) => path.resolve(__dirname, ...p);

module.exports = (ctx) => {
	const isProd = ctx.env === "production";

	const plugins = [
		postcssImport(),
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
			purgecss({
				content: [
					here("../**/*.html"),
					here("../**/*.go"),
					here("./js/**/*.ts"),
				],
				safelist: {
					standard: [
						/^is-/,
						/^has-/,
						/-active$/,
						/-open$/,
						"u-hidden",
						"dev-mode",
						"u-sr-only",
					],
					deep: [/c-realm-view\b/, /c-readme-view\b/],
				},
				variables: true,
				defaultExtractor: (content) => content.match(/[\w-:/%.]+(?<!:)/g) || [],
			}),
			cssnano({ preset: "default" }),
		);
	}

	return { plugins };
};
