# MDServe

![hero image](/assets/logo.jpg)

<a href="https://github.com/jaysongiroux/MDServe/actions/workflows/ci.yml">
  <img src="https://github.com/jaysongiroux/MDServe/actions/workflows/ci.yml/badge.svg" alt="Build Pipeline" style="height: 20px; width: auto; vertical-align: middle;">
</a>


**MDServe** is a high-performance, flat-file content server built with Go.

MDServe transforms a folder of Markdown files into a dynamic website. Unlike traditional Static Site Generators (SSGs), MDServe is a live HTTP server that can compile content on-demand or serve pre-compiled HTML from memory. Built using Go's standard library `html/template` package, it supports advanced features like custom layouts, pagination, tag filtering, and metadata-driven content organization.

It is designed to be lightweight and portable—a single binary is all you need to run your site.

## Quick Start

1. Clone the repository or download the binary
2. Create your content in the `/content` directory as Markdown files (or configure git remote content in `/config/config.yaml`)
3. Configure your site in `/config/site-config.yaml`
4. Run the server:
   ```bash
   go run main.go
   # or
   ./mdserve
   ```
5. Visit `http://localhost:8080`

## Core Features

  * **Go Standard Library:** Built using Go's `html/template` package for server-side rendering with no external framework dependencies.
  * **Configurable Rendering Engine:** Choose your performance strategy via `config.yaml`:
      * **Live Mode:** Parses Markdown on every request. Changes to content are visible immediately upon browser refresh.
      * **Static Mode:** Parses all content at startup and generates a sitemap. Serves pages with optimal performance (requires restart to update content).
  * **Git Remote Content:** Fetch content from a remote Git repository. Automatically pulls updates from configured branches and syncs content, assets, and user static files.
  * **Scheduled Content Generation:** Optional cron job for automated content regeneration at configured intervals. Perfect for keeping git-sourced content up-to-date without server restarts.
  * **Custom Layout System:** Define regex-based page matching to apply custom layouts. Perfect for blog listings, article pages, or any specialized content structure.
  * **JSON Metadata Support:** Add JSON metadata in HTML comments at the top of Markdown files with support for tags, authors, descriptions, creation dates, and custom fields.
  * **Pagination & Filtering:** Built-in support for paginated content lists with client-side tag filtering.
  * **Structure-Based Routing:** Your file system is your router. A file at `content/blog/post_1.md` is automatically served at `/blog/post_1`.
  * **Automatic Sitemap Generation:** All content is indexed and made available via a JSON sitemap for dynamic content rendering.
  * **Image Optimization:** Automatic WebP conversion and quality optimization for images in the assets folder.
  * **Cascading Styling:** Ships with a base CSS layer, but allows users to inject a `custom.css` file that automatically overrides defaults.
  * **Zero Dependencies:** Compiles into a single static binary. No external runtimes or heavy container orchestration required.


## Configuration & Architecture

MDServe separates "System Configuration" from "Site Content Configuration."

### 1. The Configuration Files

Located in the `/config` directory:

  * **`config.yaml` (System):** Controls how the server behaves.
      * Sets the port and host (e.g., `8080`, `127.0.0.1`).
      * Sets the HTML compilation mode (`live` vs `static`).
      * Defines paths for content, assets, templates, and generated files.
      * Configures image optimization settings.
      * Configures git remote content source (optional).
      * Configures generation cron for scheduled content regeneration (optional).
  * **`site-config.yaml` (Content):** Controls how the site looks and behaves.
      * Defines Navbar links with optional dropdown menus.
      * Defines Footer content and links.
      * Sets site metadata (name, powered by text).
      * Configures pagination settings (page size, sort direction).
      * Defines code syntax highlighting theme and options.
      * Maps page paths to custom layouts using regex patterns.

### 2. Configuration Management

All configuration is managed through the two YAML files in the `/config` directory. To change settings for different environments (development, production), simply modify these files before starting the server or maintain separate config files for each environment.

#### **Environment Variables:**

You can override the default config file paths using environment variables:

- **`MD_SERVER_CONFIG_PATH`**: Path or URL to the server config file (default: `config/config.yaml`)
- **`MD_SITE_CONFIG_PATH`**: Path or URL to the site config file (default: `config/site-config.yaml`)
- **`GIT_USERNAME`**: The username for git authentication.
- **`GIT_PASSWORD`**: The password or personal access token (PAT) for git authentication.

Both Config variables support:
- Local file paths (e.g., `/path/to/config.yaml`)
- HTTP/HTTPS URLs (e.g., `https://example.com/config.yaml`)

##### **Important Note for Remote Configs:**
To fetch configuration files from a remote Git repository (including private ones), you must set both/ either `MD_SERVER_CONFIG_PATH` and `MD_SITE_CONFIG_PATH` to the HTTPS Git repository URL (Ends in `.git` **Important**).

Example for remote config from a Git repository:
```bash
# Point both to the same repo URL
MD_SERVER_CONFIG_PATH="https://github.com/username/mdserve-content.git"
MD_SITE_CONFIG_PATH="https://github.com/username/mdserve-content.git"
# Optional: Specify branch (default is master/main)
MD_CONFIG_BRANCH="main"
# Optional: Specify the directory within the repo where configs are located. Default is the root of the repo
MD_CONFIG_LOCATION="config"
# Required for private repositories
GIT_USERNAME="your-username"
GIT_PASSWORD="your-pat-token"
```

The `git_remote_content_path` in `config.yaml` is exclusively for content (markdown), assets, user-static assets, and templates. It does **not** load the system configuration itself from that remote. If you want to bootstrap your server with a remote configuration, you **must** use the environment variables as shown above.

The default configuration is the `config.yaml` and `site-config.yaml` provided in this repository.

Example usage:

```bash
# Using local config files
MD_SERVER_CONFIG_PATH=/etc/mdserve/config.yaml ./mdserve

# Using remote config files
MD_SERVER_CONFIG_PATH=https://example.com/configs/server.yaml \
MD_SITE_CONFIG_PATH=https://example.com/configs/site.yaml \
./mdserve
```

### 3. Git Remote Content

MDServe supports fetching content from a remote Git repository, allowing you to host your content separately from your server. This is configured in `config.yaml`:

```yaml
# Git remote content configuration
git_remote_content_path: git@github.com:username/repo.git
git_remote_content_directory: content
git_remote_content_assets_directory: assets
git_remote_content_user_static_directory: user-static
git_remote_content_branch: master
```

#### **Configuration Options:**

- **`git_remote_content_path`**: Git URL to fetch content from (supports both HTTPS and SSH). Set to empty/null to use local content.
- **`git_remote_content_directory`**: Subdirectory in the remote repository containing Markdown content files (e.g., `content`).
- **`git_remote_content_assets_directory`**: Subdirectory in the remote repository containing assets (images, logos, icons, etc.). Set to empty/null to skip syncing remote assets.
- **`git_remote_content_user_static_directory`**: Subdirectory in the remote repository containing user static files (`custom.css`, `custom.js`, etc.). Set to empty/null to skip syncing remote user static files.
- **`git_remote_content_branch`**: Branch to fetch content from (required if `git_remote_content_path` is set).

#### **Requirements:**

- At least one directory (`git_remote_content_directory`, `git_remote_content_assets_directory`, or `git_remote_content_user_static_directory`) must be configured when using git remote content.
- Branch name is required when git remote content URL is provided.
- Git remote content and demo mode are mutually exclusive.

#### **Authentication (Private Repositories):**

To access private repositories, you must provide authentication credentials via environment variables:

- **`GIT_USERNAME`**: The username for git authentication.
- **`GIT_PASSWORD`**: The password or personal access token (PAT) for git authentication.

For Personal access tokens the minimum permissions are:
- Contents
- Metadata

> [!CAUTION]
> It is not recommended to use your password, instead use a personal access token. 


When both variables are present and a remote repository is configured in the `config.yaml`, MDServe will use Basic Auth to clone/pull from the remote repository. This is secure and recommended for private repositories.

#### **How It Works:**

When git remote content is configured:
1. On server startup, MDServe clones the remote repository to `.git-remote-content/`
2. On subsequent startups, MDServe pulls the latest changes from the specified branch
3. Configured directories are synced from the remote repository to their local counterparts
4. Local directories are cleared and replaced with remote content on each sync

This feature is useful for:
- Separating content management from server deployment
- Managing content in a separate repository
- Allowing non-technical users to edit content via Git platforms
- Deploying the same server binary with different content sources

### 4. Generation Cron

MDServe supports scheduled regeneration of content through a configurable cron job. This is particularly useful when using git remote content or when content files are updated externally.

#### **Configuration in `config.yaml`:**

```yaml
# Cron configuration
generation_cron_enabled: true
generation_cron_interval: "@hourly"
```

#### **Configuration Options:**

- **`generation_cron_enabled`**: Enable or disable the generation cron (default: `false`)
- **`generation_cron_interval`**: Cron schedule interval

#### **Supported Interval Formats:**

- **Standard cron format**: `"0 12 * * *"` (see [cron format](https://en.wikipedia.org/wiki/Cron))
- **Robfig descriptors**: `"@hourly"`, `"@daily"`, `"@weekly"`, `"@monthly"` (see [robfig/cron](https://github.com/robfig/cron))

#### **What the Cron Does:**

When triggered, the generation cron runs the preliminary setup process:
1. Pulls the latest changes from git remote content (if configured)
2. Converts Markdown files to HTML
3. Optimizes images in the assets directory
4. Generates the sitemap
5. Copies assets to the generated directory

#### **Recommended Use Case:**

When using git remote content with static compilation mode:
1. Set `html_compilation_mode: static`
2. Set `generation_cron_enabled: true`
3. Set an appropriate `generation_cron_interval` (e.g., `"@hourly"`)

When content is pushed to your remote repository, the cron will automatically pull the changes and regenerate all static files at the configured interval. This provides a balance between static site performance and relatively fresh content updates without requiring server restarts.

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
  * **`/.git-remote-content`**: Auto-generated directory when using git remote content. Contains the cloned repository.

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

## Custom Blocks
MDServe introduces some custom markdown blocks to make formatting a bit easier.

### Captions
Captions are center-aligned italic paragraphs that allow users to caption the block above it.

**Notation:**
```
^^ this is a caption ^^
^^this is also a caption^^
```


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

*Note: If your `html_compilation_mode` is set to `live`, you do not need to restart the server to see changes made to Markdown files. Simply refresh your browser. In `static` mode, you must restart the server to regenerate the sitemap and see content updates, unless you have the generation cron enabled which will automatically regenerate content at the configured interval.*


## Styling Strategy

MDServe utilizes the "Cascade" in CSS to allow for safe customization.

1.  **System Styles:** The server first loads `base.css`, which handles layout (Grid/Flexbox), typography, and responsiveness.
2.  **User Overrides:** The server checks for the existence of `custom.css` in the `/user-static` directory. If found, it is loaded *after* the base styles, allowing you to override colors, fonts, and spacing without breaking the core layout.


## Deployment

Because MDServe compiles to a single binary, deployment is flexible.

### Option 1: Standalone Binary (Recommended)

Compile the binary for your target architecture and upload it.

```bash
# Example: Build for Linux
GOOS=linux GOARCH=amd64 go build -o mdserve main.go
./mdserve
```

Ensure your `config`, `templates`, `assets`, and `user-static` folders are relative to the binary. The `content` folder is also required unless you're using git remote content. The `.static` and `.git-remote-content` directories will be auto-generated on startup when needed.

### Option 2: Docker (Optional)

A Dockerfile is provided for convenience if you prefer containerized deployment. It uses a multi-stage build to ensure a lightweight Alpine image.

#### **Build the Docker image:**

```bash
docker build -t mdserve .
```

#### **Run with default content:**

```bash
docker run -p 8080:8080 mdserve
```

#### **Run with environment variables for remote config:**

```bash
docker run -p 8080:8080 \
  -e MD_SERVER_CONFIG_PATH=https://example.com/config.yaml \
  -e MD_SITE_CONFIG_PATH=https://example.com/site-config.yaml \
  mdserve
```

#### **Run with custom content (using volumes):**

Mount your local directories to override the defaults:

```bash
docker run -p 8080:8080 \
  -v $(pwd)/content:/app/content \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/templates:/app/templates \
  -v $(pwd)/user-static:/app/user-static \
  mdserve
```

**Note:** Configuration changes require mounting the `/app/config` directory with your modified config files, or rebuilding the Docker image. If using git remote content, you don't need to mount the `/app/content` directory as it will be fetched from the remote repository.

#### **Docker Compose (recommended):**

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
