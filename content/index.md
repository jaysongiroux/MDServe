<!-- 
	NOTE: 
	This file was written because the demo mode is enabled.
	You can disable this behavior by setting the demo flag to false in the server config file.
	 -->
# MDServe

![MDServe product hero image](../assets/logo.jpg)

**MDServe** is a high-performance, flat-file content server built with Go.

MDServe transforms a folder of Markdown files into a dynamic website. Unlike traditional Static Site Generators (SSGs), MDServe is a live HTTP server that can compile content on-demand or serve pre-compiled HTML from memory. Built using Go's standard library `html/template` package, it supports advanced features like custom layouts, pagination, tag filtering, and metadata-driven content organization.

It is designed to be lightweight and portable—a single binary is all you need to run your site.

-----

## Quick Start

1. Clone the repository or download the binary
2. Create your content in the `/content` directory as Markdown files
3. Configure your site in `/config/site-config.yaml`
4. Run the server:
   ```bash
   go run main.go
   # or
   ./mdserve
   ```
5. Visit `http://localhost:8080`

-----

## Core Features

  * **Go Standard Library:** Built using Go's `html/template` package for server-side rendering with no external framework dependencies.
  * **Configurable Rendering Engine:** Choose your performance strategy via `config.yaml`:
      * **Live Mode:** Parses Markdown on every request. Changes to content are visible immediately upon browser refresh.
      * **Static Mode:** Parses all content at startup and generates a sitemap. Serves pages with optimal performance (requires restart to update content).
  * **Custom Layout System:** Define regex-based page matching to apply custom layouts. Perfect for blog listings, article pages, or any specialized content structure.
  * **JSON Metadata Support:** Add JSON metadata in HTML comments at the top of Markdown files with support for tags, authors, descriptions, creation dates, and custom fields.
  * **Pagination & Filtering:** Built-in support for paginated content lists with client-side tag filtering.
  * **Structure-Based Routing:** Your file system is your router. A file at `content/blog/post_1.md` is automatically served at `/blog/post_1`.
  * **Automatic Sitemap Generation:** All content is indexed and made available via a JSON sitemap for dynamic content rendering.
  * **Image Optimization:** Automatic WebP conversion and quality optimization for images in the assets folder.
  * **Cascading Styling:** Ships with a base CSS layer, but allows users to inject a `custom.css` file that automatically overrides defaults.
  * **Zero Dependencies:** Compiles into a single static binary. No external runtimes or heavy container orchestration required.

-----

## Configuration & Architecture

MDServe separates "System Configuration" from "Site Content Configuration."

### 1. The Configuration Files

Located in the `/config` directory:

  * **`config.yaml` (System):** Controls how the server behaves.
      * Sets the port and host (e.g., `8080`, `127.0.0.1`).
      * Sets the HTML compilation mode (`live` vs `static`).
      * Defines paths for content, assets, templates, and generated files.
      * Configures image optimization settings.
  * **`site-config.yaml` (Content):** Controls how the site looks and behaves.
      * Defines Navbar links with optional dropdown menus.
      * Defines Footer content and links.
      * Sets site metadata (name, powered by text).
      * Configures pagination settings (page size, sort direction).
      * Defines code syntax highlighting theme and options.
      * Maps page paths to custom layouts using regex patterns.

### 2. Configuration Management

All configuration is managed through the two YAML files in the `/config` directory. To change settings for different environments (development, production), simply modify these files before starting the server or maintain separate config files for each environment.

-----

## Directory Structure

The system expects a specific folder hierarchy to function:

  * **`/content`**: The heart of the site. Markdown files here become URLs. Subdirectories become route paths.
  * **`/config`**:
      * `config.yaml`: System settings.
      * `site-config.yaml`: Site content, navigation, and layout definitions.
  * **`/templates`**: HTML layouts using Go Templates.
      * `layout.html`: Master page layout.
      * `navbar.html`, `footer.html`, `scripts.html`: Partial templates.
      * `/layout_templates`: Custom page layouts (e.g., blog listing, article pages).
  * **`/assets`**: System-level static files (images, base CSS, base JS).
  * **`/user-static`**: User-provided assets (like `custom.css` and `custom.js`) that persist across updates.
  * **`/.static`**: Auto-generated files (sitemap.json) created at startup in static mode.

-----

## Markdown Metadata

MDServe supports JSON metadata embedded in HTML comments at the top of Markdown files:

```markdown
<!-- 
{
    "tags": ["golang", "web development"],
    "creation_date": "2025-01-15T14:00:00Z",
    "last_modification_date": "2025-02-01T16:00:00Z",
    "author": "John Doe",
    "description": "A brief description of the post that overrides the first paragraph in excerpts"
}
-->

# Your Markdown Content Here
```

Metadata fields are automatically extracted and made available in custom layouts via the sitemap. If `creation_date` or `last_modification_date` are not provided, the file's modification time is used as a fallback.

-----

## Custom Layouts

Custom layouts allow you to create specialized templates for different sections of your site. Define layouts in `site-config.yaml`:

```yaml
layouts:
  # Match blog article pages
  - page: blog\/posts\/.*
    layout: blog_article_layout
  
  # Match the blog index with filtering
  - page: ^blog$
    filter: blog/posts/.*
    layout: blog_layout
```

Layout templates are stored in `/templates/layout_templates/` and have access to:
- `{{ .Content }}`: The compiled HTML content
- `{{ .SiteMap }}`: Full sitemap of all pages
- `{{ .PageList }}`: Filtered list of pages (when `filter` is specified)
- `{{ .Site }}`: Site configuration (page_size, theme, etc.)

-----

## Development Workflow

MDServe is written in Go. For the best development experience, we recommend using **Air** for live reloading.

### Prerequisites

  * Go 1.22+
  * [Air](https://github.com/cosmtrek/air) (Optional, for hot reloading binary)

### Running Locally

1.  **Standard Run:**
    ```bash
    go run main.go
    ```
2.  **With Live Reload (Air):**
    If you are modifying the **Go source code** or **HTML templates**, run `air` in the root directory. This will watch your `.go` files and automatically rebuild/restart the binary when code changes.
    ```bash
    air
    ```

*Note: If your `html_compilation_mode` is set to `live`, you do not need to restart the server to see changes made to Markdown files. Simply refresh your browser. In `static` mode, you must restart the server to regenerate the sitemap and see content updates.*

-----

## Styling Strategy

MDServe utilizes the "Cascade" in CSS to allow for safe customization.

1.  **System Styles:** The server first loads `base.css`, which handles layout (Grid/Flexbox), typography, and responsiveness.
2.  **User Overrides:** The server checks for the existence of `custom.css` in the `/user-static` directory. If found, it is loaded *after* the base styles, allowing you to override colors, fonts, and spacing without breaking the core layout.

-----

## Deployment

Because MDServe compiles to a single binary, deployment is flexible.

### Option 1: Standalone Binary (Recommended)

Compile the binary for your target architecture and upload it.

```bash
# Example: Build for Linux
GOOS=linux GOARCH=amd64 go build -o mdserve main.go
./mdserve
```

Ensure your `config`, `content`, `templates`, `assets`, and `user-static` folders are relative to the binary. The `.static` directory will be auto-generated on startup when running in static mode.

### Option 2: Docker (Optional)

A Dockerfile is provided for convenience if you prefer containerized deployment. It uses a multi-stage build to ensure a lightweight Alpine image.

#### Build the Docker image:

```bash
docker build -t mdserve .
```

#### Run with default content:

```bash
docker run -p 8080:8080 mdserve
```

#### Run with custom content (using volumes):

Mount your local directories to override the defaults:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/content:/app/content \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/templates:/app/templates \
  -v $(pwd)/user-static:/app/user-static \
  mdserve
```

**Note:** Configuration changes require mounting the `/app/config` directory with your modified config files, or rebuilding the Docker image.

#### Docker Compose (recommended):

A `docker-compose.yml` file is included in the project. Simply run:

```bash
docker-compose up
```

Or to run in detached mode:

```bash
docker-compose up -d
```

To stop the container:

```bash
docker-compose down
```
