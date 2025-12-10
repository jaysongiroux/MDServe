package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/jaysongiroux/mdserve/internal/auth"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

func GetFileFromRepo(
	url string,
	branch *string,
	path string,
	configFileName string,
) (string, error) {
	fs := memfs.New()

	var branchName plumbing.ReferenceName
	if branch != nil && *branch != "" {
		branchName = plumbing.NewBranchReferenceName(*branch)
	} else {
		branchName = ""
	}

	username := os.Getenv(ENV_VAR_GIT_USERNAME)
	password := os.Getenv(ENV_VAR_GIT_PASSWORD)

	_, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:           url,
		ReferenceName: branchName,
		SingleBranch:  true,
		Depth:         1, // shallow clone (no history)
		Auth:          auth.CreateGitBasicAuth(&username, &password),
	})
	if err != nil {
		logger.Error("Failed to clone git repository: %v", err)
		return "", fmt.Errorf("failed to clone git repository: %w", err)
	}

	content, err := fs.Open(filepath.Join(path, configFileName))
	if err != nil {
		logger.Error("Failed to open file: %v. %v", filepath.Join(path, configFileName), err)
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := content.Close()
		if err != nil {
			logger.Error("Failed to close file: %v", err)
		}
	}()

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		logger.Error("Failed to read file: %v", err)
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	logger.Debug("Successfully read file: %s", path)
	return string(contentBytes), nil
}
