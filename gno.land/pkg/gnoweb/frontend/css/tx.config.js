const pxToRem = (px) => px / 16;

export default {
  content: ["./components/**/*.{gohtml,ts}"],
  theme: {
    screens: {
      xs: `${pxToRem(360)}rem`,
      sm: `${pxToRem(480)}rem`,
      md: `${pxToRem(640)}rem`,
      lg: `${pxToRem(820)}rem`,
      xl: `${pxToRem(1020)}rem`,
      xxl: `${pxToRem(1366)}rem`,
      max: `${pxToRem(1580)}rem`,
    },
    zIndex: {
      min: "-1",
      1: "1",
      2: "2",
      100: "100",
      max: "9999",
    },
    container: {
      center: true,
      padding: `${pxToRem(40)}rem`,
    },
    borderRadius: {
      sm: `${pxToRem(4)}rem`,
      DEFAULT: `${pxToRem(6)}rem`,
    },
    colors: {
      light: "#FFFFFF",
      gray: {
        50: "#F0F0F0", // Background color
        100: "#E2E2E2", // Title dark color
        200: "#BDBDBD", // Content dark color
        300: "#999999", // Muted color
        400: "#7C7C7C", // Border color
        600: "#54595D", // Content color
        800: "#131313", // Background dark color
        900: "#080809", // Title color
      },
      green: {
        400: "#2D8D72", // Primary dark color
        600: "#226C57", // Primary light color
      },
      transparent: "transparent",
      current: "currentColor",
      inherit: "inherit",
    },
    fontFamily: {
      mono: ["Roboto", 'Menlo, Consolas, "Ubuntu Mono", "Roboto Mono", "DejaVu Sans Mono", monospace;'],
      interVar: [
        '"Inter var"',
        '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji", sans-serif',
      ],
    },
    fontSize: {
      0: "0",
      50: `${pxToRem(12)}rem`,
      100: `${pxToRem(14)}rem`,
      200: `${pxToRem(16)}rem`,
      300: `${pxToRem(18)}rem`,
      400: `${pxToRem(20)}rem`,
      500: `${pxToRem(22)}rem`,
      600: `${pxToRem(24)}rem`,
      700: `${pxToRem(32)}rem`,
      800: `${pxToRem(38)}rem`,
      900: `${pxToRem(42)}rem`,
    },
  },
  plugins: [],
};
