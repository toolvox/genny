# genny

Static site generator that combines HTML templates, YAML data, and reusable components.

## Usage

```bash
genny [flags] [path]    # defaults to current directory
genny -w [path]         # watch mode: automatically regenerate on file changes
genny -v [path]         # verbose mode: show detailed logging
```

### Flags
- `-w`, `-watch` - Watch for file changes and regenerate automatically
- `-v`, `-verbose` - Enable verbose logging
- `-h`, `-help` - Show help message

## Project Structure

```
./
├── assets/          # Static assets (images, fonts, etc.)
├── data/            # YAML data files (*.yaml)
├── components/      # Reusable HTML components (*.html)
├── *.html           # Project pages at root level (e.g., paz.html, google.html)
├── */index.html     # Alternative: Project pages in subdirectories (for backward compatibility)
├── index.html       # Main template with component references
├── header.html      # Header for every generated page
├── footer.html      # Footer for every generated page
├── decrypt.html     # Decrypt form template for encrypted pages (auto-created if needed)
├── *.css            # Stylesheets (all CSS files are copied to output)
└── www/             # Generated output directory
    ├── index.html   # Main site
    ├── *.html       # Generated project pages (flat structure)
    ├── */index.html # Generated project pages (subdirectory structure)
    ├── preview/     # Component and page previews
    ├── assets/      # Copied static assets
    └── *.css        # Copied stylesheets
```

## How It Works

1. **Loads runtime data:**
   - Assets from `./assets/`
   - YAML data from `./data/` (each file namespaced by filename, e.g., `data/projects.yaml` → `.projects`)
   - HTML components from `./components/`
   - Page files:
     - `.html` files at root level (except `index.html`, `header.html`, `footer.html`)
     - `index.html` files in subdirectories (backward compatibility)

2. **Processes templates:**
   - Extracts wrapper from `index.html` (splits on `<body>` tag)
   - Component files define their data path via `<preview>` tag in `<head>`
   - Automatically prepends `.` to data paths if missing (e.g., `DataPath` becomes `.DataPath`)
   - Wraps all pages with header and footer templates
   - Components can reference other components using `<component_name>` tags
   - Converts component tags to Go template syntax: `<foo>.Key.To.Data</foo>` → `{{ template "foo" .Key.To.Data }}`

3. **Generates output:**
   - Main site from root `index.html` template
   - All discovered page files (flat and subdirectory)
   - Encrypted pages: pages with `<encrypt>` tags are AES-256-GCM encrypted; output contains decrypt form + encrypted payload
   - Component previews (wrapped in index.html structure)
   - Page previews (including index, in `www/preview/`) - always unencrypted
   - Executes templates with full YAML data context
   - Applies whitespace cleanup to remove excessive newlines
   - Adjusts asset/stylesheet paths for directory depth (and for preview directory)
   - Copies all assets and stylesheets (`*.css`) to `./www/`

## Component Files

Components specify their data source using a `<preview>` tag in the `<head>` section:

```html
<!DOCTYPE HTML>
<html>
<head>
    <preview>DataPath.To.Object</preview>
</head>
<body>
    <div>{{ .SomeField }}</div>
    <img src="{{ .ImageURL }}" alt="{{ .AltText }}">
</body>
</html>
```

The `<preview>` tag specifies the path in the YAML data to use for rendering this component. If the path doesn't start with `.`, it will be automatically prepended (e.g., `DataPath.To.Object` becomes `.DataPath.To.Object`).

## Page Files

Page files can be structured in two ways:
1. **Flat structure (recommended):** `.html` files at root level (e.g., `paz.html`, `google.html`, `leumi.html`)
2. **Subdirectory structure:** `index.html` files in subdirectories (e.g., `paz/index.html`)

Both structures are supported for backward compatibility. Pages in subdirectories excluding `components/`, `data/`, `assets/`, and `www/` are auto-discovered. These pages:
- Are automatically wrapped with `header.html` and `footer.html` templates
- Have asset and stylesheet paths adjusted based on their directory depth (0 for flat files)
- Support component tags just like the main `index.html`
- Are regenerated when modified in watch mode

## Encrypted Pages

Pages can be password-protected by adding an `<encrypt>` tag in the `<head>` section:

```html
<!doctype html>
<html>
<head>
    <title>Secret Page</title>
    <encrypt>my-passphrase</encrypt>
</head>
<body>
    <h1>This content will be encrypted</h1>
</body>
</html>
```

When a page has an `<encrypt>` tag:
- The **main output** (`www/{page}.html`) is encrypted with AES-256-GCM (key derived via PBKDF2-SHA256). The generated file contains a password form and inline JavaScript that decrypts the page in the browser when the correct passphrase is entered.
- The **preview** (`www/preview/{page}.html`) is generated normally, fully unencrypted, for development use.
- The `<encrypt>` tag is stripped from all output.

The decrypt form UI comes from `decrypt.html` at the site root. If this file doesn't exist when an encrypted page is first encountered, a default one is auto-created. You can customize it like any other site-level template (`<html>/<head>/<body>` structure) - only the `<body>` content is used.

## Data Flow

YAML files in `./data/` are loaded and namespaced by filename (e.g., `data/projects.yaml` is accessible as `.projects` in templates). Components are matched to their data via the `<preview>` data path specification, then rendered using Go's `html/template` package.

---

## Architecture

The codebase is organized into focused packages with clear responsibilities:

```
pkg/
├── cli/              - Command-line interface argument parsing
│   └── cli.go        - Flag parsing and config management
├── encrypt/          - Page encryption for password-protected pages
│   ├── encrypt.go    - AES-256-GCM encryption with PBKDF2-SHA256 key derivation
│   └── decrypt_template.go - Decrypt page HTML template with inline WebCrypto JS
├── generator/        - Site generation logic
│   ├── types.go      - Core domain types (Site, Component, Page, Asset)
│   ├── errors.go     - Custom error types
│   ├── component_generator.go - Component preview generation
│   ├── site_generator.go      - Main site and page preview generation
│   └── path_adjuster.go       - Path adjustment for output
├── loader/           - File I/O operations
│   ├── loader.go     - Loader interface
│   ├── assets.go     - Asset discovery and loading
│   ├── data.go       - YAML data file loading
│   ├── components.go - Component file discovery
│   ├── pages.go      - Page discovery (root-level .html and subdirectory index.html)
│   └── templates.go  - Template file loading
├── orchestrator/     - Workflow coordination
│   └── orchestrator.go - RunOnce and RunContinuous modes
├── parser/           - HTML and template parsing
│   ├── component_parser.go - Extract data paths from components
│   └── tag_replacer.go     - Convert component tags to template syntax
├── site/             - High-level site orchestration
│   └── site.go       - Coordinates loading, parsing, and generation
├── utils/            - Utility functions
│   └── html.go       - HTML parsing utilities and whitespace cleanup
└── watcher/          - File system monitoring
    └── watcher.go    - fsnotify-based file watching with debouncing
```
