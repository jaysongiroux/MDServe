# MDServe Quick Start

This guide provides a structured approach to setting up MDServe for different use cases.

## 1. Using This Repository as a Template

If you are just getting started, you can use this repository as-is. It comes pre-configured with:
*   **Templates**: Default HTML layouts.
*   **Config**: Up-to-date `server_config.yaml` and `site_config.yaml`.
*   **Content**: Example markdown files.
*   **Assets**: Default styling and scripts.

### Customization Steps
1.  **Content**: Add, rename, or delete Markdown files in the `/content` directory.
2.  **Configuration**: Modify `/config/site-config.yaml` to update navigation and site metadata.
3.  **Styles**: Add a `custom.css` file to `/user-static` to override default styles.
4.  **Run**: Execute `go run main.go` or `./mdserve`.

---

## 2. Using Remote Content (GitOps)

This approach allows you to decouple your content from the server deployment. You manage your content in a separate Git repository, and MDServe pulls it automatically.

### Recommended Remote Repository Structure
Create a dedicated repository (e.g., `my-blog-content`) with the following structure:
```text
/
├── assets/         # Images, icons
├── content/        # Markdown files
├── config/         # Optional: Remote config files
├── user-static/    # custom.css, custom.js
└── templates/      # Optional: Custom templates
```

### Setup Process
1.  **Populate Repo**: Copy the contents of the respective directories from the MDServe repo to your new repository.
2.  **Push**: Commit and push changes to your remote repository.
3.  **Configure Environment**: Point your server to the remote configuration.

```bash
# Example .env configuration
# Point both to the same repo URL
MD_SERVER_CONFIG_PATH="https://github.com/jaysongiroux/mdserve-content.git"
MD_SITE_CONFIG_PATH="https://github.com/jaysongiroux/mdserve-content.git"

# Optional: Specify branch (default is master/main)
MD_CONFIG_BRANCH="main"

# If using a private repository
GIT_USERNAME=your_username
GIT_PASSWORD=your_pat_token
```

> **Note**: The `git_remote_content_path` setting in `config.yaml` is only for content. To load the **configuration itself** from a remote repo, you must use the environment variables above. The default config is the one provided in this repo.

---

## 3. Using Local Directories (External Paths)

Use this method if you want to store your content, templates, and assets in a specific location on your local filesystem, outside the application directory.

### Setup Process
1.  **Prepare Directories**: Create your external folder structure (e.g., `/var/www/mdserve/`).
    ```text
    /var/www/mdserve/
    ├── content/
    ├── assets/
    ├── templates/
    └── user-static/
    ```
2.  **Migrate Files**: Copy the corresponding files from the MDServe repo to these directories.
3.  **Update Config**: Modify your `config.yaml` to point to absolute paths.

**Example `config.yaml` adjustment:**
```yaml
content_path: /var/www/mdserve/content
assets_path: /var/www/mdserve/assets
templates_path: /var/www/mdserve/templates
user_static_path: /var/www/mdserve/user-static
```

4.  **Run**: Start MDServe, and it will serve content from the specified absolute paths.
