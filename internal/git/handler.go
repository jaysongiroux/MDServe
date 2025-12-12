package git

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

func HandleSyncFromRepo(serverConfig *config.ServerConfig) error {
	if !serverConfig.SyncAssets && !serverConfig.SyncTemplates {
		logger.Debug("Syncing templates and assets is disabled, skipping")
		return nil
	}

	fs, err := tempCloneRepo()
	if err != nil {
		logger.Error("Failed to clone repo: %v", err)
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	if serverConfig.SyncTemplates {
		logger.Debug("Syncing templates is enabled, syncing templates")
		if err := syncTemplates(serverConfig, fs); err != nil {
			return fmt.Errorf("failed to sync templates: %w", err)
		}
	}

	if serverConfig.SyncAssets {
		logger.Debug("Syncing assets is enabled, syncing assets")
		if err := syncAssets(serverConfig, fs); err != nil {
			return fmt.Errorf("failed to sync assets: %w", err)
		}
	}

	return nil
}

func tempCloneRepo() (billy.Filesystem, error) {
	fs := memfs.New()

	_, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:          constants.RepoUrl,
		SingleBranch: true,
		Depth:        1,
	})
	if err != nil {
		logger.Error("Failed to clone repo: %v", err)
		return nil, fmt.Errorf("failed to clone repo: %w", err)
	}

	return fs, nil
}

func syncAssets(serverConfig *config.ServerConfig, fs billy.Filesystem) error {
	logger.Info("Syncing assets from remote repository")
	return syncMemFSFilesystemDirectory(fs, constants.RepoAssetsDirectory, serverConfig.AssetsPath)
}

func syncTemplates(serverConfig *config.ServerConfig, fs billy.Filesystem) error {
	logger.Info("Syncing templates from remote repository")
	return syncMemFSFilesystemDirectory(fs, constants.RepoTemplatesDirectory, serverConfig.TemplatesPath)
}

func syncMemFSFilesystemDirectory(fs billy.Filesystem, remoteDir string, localDir string) error {
	// Get all files from remote directory
	remoteFiles, err := getAllFilesInFS(fs, remoteDir)
	if err != nil {
		return fmt.Errorf("failed to read remote directory %s: %w", remoteDir, err)
	}

	logger.Debug("Found %d files in remote directory %s", len(remoteFiles), remoteDir)

	// Ensure local directory exists
	if err := os.MkdirAll(localDir, 0750); err != nil {
		return fmt.Errorf("failed to create local directory %s: %w", localDir, err)
	}

	// Sync each file
	for _, remoteFile := range remoteFiles {
		// Calculate relative path from remote directory
		relPath := strings.TrimPrefix(remoteFile, remoteDir)
		relPath = strings.TrimPrefix(relPath, "/")

		localPath := filepath.Join(localDir, relPath)

		if err := syncFile(fs, remoteFile, localPath); err != nil {
			return fmt.Errorf("failed to sync file %s: %w", remoteFile, err)
		}
	}

	logger.Info("Successfully synced directory %s to %s", remoteDir, localDir)
	return nil
}

func syncFile(fs billy.Filesystem, remotePath, localPath string) error {
	// Read remote file
	remoteFile, err := fs.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer func() {
		err := remoteFile.Close()
		if err != nil {
			logger.Error("Failed to close remote file: %v", err)
		}
	}()

	remoteContent, err := io.ReadAll(remoteFile)
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
	}

	// Calculate remote file hash
	remoteHash := calculateHash(remoteContent)

	// Check if local file exists
	localExists := true
	localContent, err := os.ReadFile(filepath.Clean(localPath))
	if err != nil {
		if os.IsNotExist(err) {
			localExists = false
		} else {
			return fmt.Errorf("failed to read local file: %w", err)
		}
	}

	// Determine if we need to update
	shouldUpdate := false
	if !localExists {
		logger.Debug("Local file %s does not exist, creating new file", localPath)
		shouldUpdate = true
	} else {
		localHash := calculateHash(localContent)
		if remoteHash != localHash {
			logger.Warn("File hash mismatch for %s - overwriting local file with remote version", localPath)
			logger.Debug("Local hash: %s, Remote hash: %s", localHash, remoteHash)
			shouldUpdate = true
		} else {
			logger.Debug("File %s is up to date, skipping", localPath)
		}
	}

	// Write file if needed
	if shouldUpdate {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		if err := os.WriteFile(localPath, remoteContent, 0600); err != nil {
			return fmt.Errorf("failed to write local file: %w", err)
		}

		if localExists {
			logger.Info("Overwrote local file: %s", localPath)
		} else {
			logger.Info("Created new file: %s", localPath)
		}
	}

	return nil
}

func getAllFilesInFS(fs billy.Filesystem, dir string) ([]string, error) {
	var files []string

	// Read directory
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively get files from subdirectories
			subFiles, err := getAllFilesInFS(fs, fullPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, fullPath)
		}
	}

	return files, nil
}

func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
