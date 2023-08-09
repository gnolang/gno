// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
    title: 'Gno.land Documentation',
    favicon: 'img/favicon.ico',
    url: 'https://docs.gno.land',
    baseUrl: '/',

    organizationName: 'gnolang',
    projectName: 'gno',

    onBrokenLinks: 'throw',
    onBrokenMarkdownLinks: 'warn',

    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    presets: [
        [
            'classic',
            /** @type {import('@docusaurus/preset-classic').Options} */
            ({
                docs: {
                    routeBasePath: '/',
                    sidebarPath: require.resolve('./sidebars.js'),
                },
                blog: false,
                theme: {
                    customCss: require.resolve('./src/css/custom.css'),
                },
            }),
        ],
    ],

    themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
        ({
            navbar: {
                hideOnScroll: true,
                title: 'Gno.land',
                logo: {
                    alt: 'Gno.land Logo',
                    src: 'img/logo.svg',
                    srcDark: 'img/logo_light.svg'
                },
                items: [
                    {
                        type: 'docSidebar',
                        sidebarId: 'tutorialSidebar',
                        position: 'left',
                        label: 'Docs',
                    },
                    {
                        href: 'https://github.com/gnolang/gno',
                        label: 'GitHub',
                        position: 'right',
                    },
                ],
            },
            footer: {
                style: 'dark',
                links: [
                    {
                        title: 'Socials',
                        items: [
                            {
                                label: 'Discord',
                                href: 'https://discord.gg/S8nKUqwkPn',
                            },
                            {
                                label: 'Twitter',
                                href: 'https://twitter.com/_gnoland',
                            },
                            {
                                label: 'YouTube',
                                href: 'https://www.youtube.com/@_gnoland',
                            },
                            {
                                label: 'Telegram',
                                href: 'https://t.me/gnoland',
                            },
                        ],
                    },
                    {
                        title: 'Gno Libraries',
                        items: [
                            {
                                label: 'gno-js-client',
                                href: 'https://github.com/gnolang/gno-js-client',
                            },
                            {
                                label: 'tm2-js-client',
                                href: 'https://github.com/gnolang/tm2-js-client',
                            },
                        ],
                    },
                ],
                copyright: `Made with ❤️ by the humans at <a href='https://gno.land'>Gno.land</a>`,
            },
            prism: {
                theme: lightCodeTheme,
                darkTheme: darkCodeTheme,
            },
        }),
};

module.exports = config;
