---
name: genny-site
description: Building static websites with the genny static site generator. Use when creating pages, components, data files, or troubleshooting genny site projects.
user-invocable: true
---

# Building Sites with Genny

Genny generates static sites from HTML templates, YAML data, and reusable components.

## CLI

```bash
genny [path]        # generate site (defaults to current directory)
genny -w [path]     # watch mode: regenerate on file changes
genny -v [path]     # verbose mode: detailed logging
```

## Project Structure

```
./
├── index.html       # Main page template (required)
├── header.html      # Header included on every page (optional)
├── footer.html      # Footer included on every page (optional)
├── decrypt.html     # Decrypt form for encrypted pages (auto-created if needed)
├── *.css            # Stylesheets (all copied to output)
├── assets/          # Static files: images, fonts, etc.
├── data/            # YAML data files
├── components/      # Reusable HTML components
├── *.html           # Additional pages (e.g., about.html, contact.html)
├── */index.html     # Alternative: pages in subdirectories
└── www/             # Generated output (don't edit)
    ├── index.html
    ├── *.html
    ├── preview/     # Auto-generated previews of components and pages
    ├── assets/
    └── *.css
```

## Data

YAML files in `data/` are namespaced by filename and accessible in templates:
- `data/projects.yaml` -> `.projects` in templates
- `data/site.yaml` -> `.site` in templates

## Components

Components live in `components/` as `.html` files. A component has:
- A `<preview>` tag in `<head>` specifying which YAML data path to use for its standalone preview
- A `<body>` containing the template, using Go template syntax (`{{ .Field }}`)

```html
<!DOCTYPE HTML>
<html>
<head>
    <preview>projects.Featured</preview>
</head>
<body>
    <div class="card">
        <h2>{{ .Title }}</h2>
        <p>{{ .Description }}</p>
    </div>
</body>
</html>
```

The `<preview>` path points into the YAML data. A leading `.` is added automatically if missing.

### Using Components

Reference components in any page or template using custom HTML tags. The tag name matches the component filename (without `.html`). The content between tags is the data path:

```html
<project_card>.projects.Featured</project_card>
```

Components can nest other components the same way.

## Pages

Pages are standard HTML files with `<head>` and `<body>` tags. They are automatically wrapped with `header.html` and `footer.html`, and can use component tags and Go template syntax.

Two layouts are supported:
1. **Flat (recommended):** `about.html`, `contact.html` at root level
2. **Subdirectory:** `about/index.html` (for backward compatibility)

Pages have access to all YAML data via the same dot paths as components.

## Previews

Genny auto-generates standalone preview files in `www/preview/` for every component and every page. These are useful for reviewing individual pieces in isolation.

## Encrypted Pages

Add an `<encrypt>` tag in a page's `<head>` to password-protect it:

```html
<head>
    <title>Secret Page</title>
    <encrypt>my-passphrase</encrypt>
</head>
```

- Main output (`www/{page}.html`) is AES-256-GCM encrypted with a password form for browser-side decryption
- Preview (`www/preview/{page}.html`) is always unencrypted
- `decrypt.html` at the site root provides the form UI (auto-created with a default if missing, customizable)
- The `<encrypt>` tag is stripped from all output

## Output

All generated files go to `www/`. Asset and stylesheet paths are automatically adjusted for directory depth. Genny warns about components that aren't referenced anywhere.
