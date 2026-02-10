# GnoWeb Frontend

This directory contains the frontend assets for GnoWeb, including CSS and JavaScript build systems.

## Architecture

### CSS System
- **ITCSS + CUBE CSS**: Logical CSS organization with Composition, Utility, Block, Exception
- **CSS Custom Properties**: Token-based system for consistent values
- **PostCSS**: Modern CSS processing with autoprefixer and optimizations

### JavaScript System
- **esbuild**: Fast JavaScript bundling and minification
- **TypeScript**: Type-safe JavaScript development
- **Controller Base Extension**: Extends base controllers with additional functionality
- **Event Handling**: DOM event management and delegation
- **Modular structure**: Organized JavaScript modules

## File Structure

```
css/
├── tokens.css          # CSS variables (auto-generated)
├── 01-settings.css     # Configuration and switches
├── 02-tools.css        # Mixins and functions
├── 03-generic.css      # Reset and base styles
├── 04-elements.css     # HTML element styles
├── 05-composition.css  # Layout patterns
├── 06-blocks.css       # Reusable components
├── 07-utilities.css    # Utility classes
└── main.css            # Main entry point

js/
├── index.ts            # Main JavaScript entry point
└── controller-[foo]    # JavaScript controller modules (prefixed)

static/
├── fonts/              # Font files
└── imgs/               # Image files
```

## Build Commands

### NPM Scripts
```bash
# Install dependencies
npm install

# Build
TODO
```

### Go Scripts
```bash
# Generate CSS tokens
TODO
```

## Configuration

### CSS Tokens
CSS tokens are defined in `TODO` and automatically generated into `TODO`. This includes:
- Colors and themes
- Spacing and typography
- Breakpoints and shadows
- Border radius and transitions

### JavaScript Configuration
JavaScript is configured through esbuild with TypeScript support. The build process:
- Bundles all modules into a single file
- Minifies for production
- Generates source maps and logs for development

The JavaScript system uses a controller-based architecture that extends base controllers to provide:
- Enhanced DOM manipulation capabilities
- Event delegation and handling
- Simple Component lifecycle management and synchronization

## Development Workflow

1. **CSS Development**: Edit CSS files in the `css/` directory
2. **JavaScript Development**: Edit TypeScript files in the `js/` directory

## Customization

### CSS Tokens
Modify `TODO` to change design tokens, then run:
```bash
go run TODO
```

### Theme Support
The system supports dark/light themes via CSS custom properties and `[data-theme="dark"]` selector.

## Build Process

### CSS Pipeline
1. CSS files are processed through PostCSS
2. Tokens are generated from configuration
3. Files are concatenated in ITCSS order
4. Output is optimized and minified

### JavaScript Pipeline
1. TypeScript files are compiled
2. Modules are bundled with esbuild
3. Output is minified for production
4. Source maps are generated for development

## Dependencies

### CSS Dependencies
- PostCSS and plugins for CSS processing
- TODO for browser compatibility
- TODO for optimization

### JavaScript Dependencies
- esbuild for bundling
- TypeScript for type safety
