package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/jaysongiroux/mdserve/internal/auth"
	"github.com/jaysongiroux/mdserve/internal/config"
	"github.com/jaysongiroux/mdserve/internal/constants"
	"github.com/jaysongiroux/mdserve/internal/files"
	"github.com/jaysongiroux/mdserve/internal/logger"
)

// Errors
const (
	ErrAlreadyUpToDate = "already up-to-date"
	ErrObjectNotFound  = "object not found"
	ErrOutOfSync       = "out of sync"
)

func fetchGitRemoteContent(url string, destinationPath string, branch string) error {
	// clone the remote repository to the destination path
	logger.Debug("Cloning git remote content from %s to %s", url, destinationPath)
	username := os.Getenv(config.ENV_VAR_GIT_USERNAME)
	password := os.Getenv(config.ENV_VAR_GIT_PASSWORD)
	repo, err := git.PlainClone(destinationPath, &git.CloneOptions{
		URL:          url,
		Depth:        1,
		SingleBranch: true,
		Auth:         auth.CreateGitBasicAuth(&username, &password),
	})

	if err != nil {
		logger.Error("Failed to clone git remote content from %s to %s: %v", url, destinationPath, err)
		return err
	}

	branchRef, err := repo.Branch(branch)
	if err != nil {
		logger.Error("Failed to get branch %s: %v", branch, err)
		return err
	}
	if branchRef == nil {
		logger.Error("Branch %s does not exist", branch)
		return fmt.Errorf("branch %s does not exist", branch)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		logger.Error("Failed to get worktree: %v", err)
		return err
	}

	logger.Debug("Checking out branch %s", branch)
	branchRefName := plumbing.NewBranchReferenceName(branch)
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
		Force:  true,
	})
	if err != nil {
		logger.Error("Failed to checkout branch %s: %v", branch, err)
		return err
	}

	logger.Debug("Successfully checked out branch %s", branch)

	return nil
}

func openGitRepository(directory string) (*git.Repository, error) {
	logger.Debug("Opening git repository at %s", directory)
	repo, err := git.PlainOpen(directory)
	if err != nil {
		return nil, err
	}
	logger.Debug("Successfully opened git repository at %s", directory)
	return repo, nil
}

func createGitRemoteContentDirectory() (string, error) {
	// create a directory to clone the git remote content to if it doesnt exist
	logger.Debug("Creating git remote content directory at %s", constants.GitRemoteContentDirectory)
	err := os.MkdirAll(constants.GitRemoteContentDirectory, 0750)
	if err != nil {
		return "", err
	}

	logger.Debug("Successfully created git remote content directory at %s", constants.GitRemoteContentDirectory)
	return constants.GitRemoteContentDirectory, nil
}

func pullLatestGitRemoteContent(branch string, directory string) error {
	username := os.Getenv(config.ENV_VAR_GIT_USERNAME)
	password := os.Getenv(config.ENV_VAR_GIT_PASSWORD)

	// open the repository
	repo, err := openGitRepository(directory)
	if err != nil {
		logger.Warn("Failed to open git repository at %s: %v", directory, err)
		return fmt.Errorf("git repository is %s: %w", ErrOutOfSync, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		logger.Warn("Failed to get worktree: %v", err)
		return fmt.Errorf("git repository is %s: %w", ErrOutOfSync, err)
	}

	logger.Debug("Pulling the latest changes from branch %s", branch)
	branchRefName := plumbing.NewBranchReferenceName(branch)
	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: branchRefName,
		Force:         true,
		Auth:          auth.CreateGitBasicAuth(&username, &password),
	})

	if err != nil {
		if strings.Contains(err.Error(), ErrAlreadyUpToDate) {
			logger.Debug("Branch %s is already up-to-date", branch)
		} else if strings.Contains(err.Error(), ErrObjectNotFound) {
			// Object not found error typically occurs with shallow clones when
			// the remote has been force-pushed or rebased, or when the repository
			// is corrupted (common in Docker container restarts without volumes)
			logger.Warn("Git objects not found for branch %s, repository needs to be re-cloned", branch)
			return fmt.Errorf("git repository is %s: %w", ErrOutOfSync, err)
		} else if strings.Contains(err.Error(), "reference not found") || 
		          strings.Contains(err.Error(), "couldn't find remote ref") ||
		          strings.Contains(err.Error(), "repository does not exist") {
			// These errors indicate repository corruption or invalid state
			logger.Warn("Repository reference errors detected, repository needs to be re-cloned")
			return fmt.Errorf("git repository is %s: %w", ErrOutOfSync, err)
		} else {
			logger.Error("Failed to pull down the latest changes from branch %s: %v", branch, err)
			return err
		}
	} else {
		logger.Debug("Successfully pulled down the latest changes from branch %s", branch)
	}

	return nil
}

func isGitRemoteContentDirectoryAGitRepository() (bool, error) {
	repo, err := openGitRepository(constants.GitRemoteContentDirectory)
	if err != nil {
		return false, nil
	}
	
	// Validate that the repository has a valid worktree
	// This helps detect corrupted or incomplete repositories (common in Docker restarts)
	worktree, err := repo.Worktree()
	if err != nil {
		logger.Debug("Repository exists but worktree is invalid: %v", err)
		return false, nil
	}
	
	// Validate that the repository has a valid HEAD reference
	// This ensures the repository has been properly initialized with at least one commit
	_, err = repo.Head()
	if err != nil {
		logger.Debug("Repository exists but HEAD reference is invalid: %v", err)
		return false, nil
	}
	
	// Check if the worktree directory exists and is not empty
	// In ephemeral Docker containers, the directory might exist but be empty
	status, err := worktree.Status()
	if err != nil {
		logger.Debug("Repository exists but status check failed: %v", err)
		return false, nil
	}
	
	// Additional validation: check if .git/config exists
	// This is a critical file that should always exist in a valid git repository
	gitConfigPath := filepath.Join(constants.GitRemoteContentDirectory, ".git", "config")
	if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
		logger.Debug("Repository directory exists but .git/config is missing")
		return false, nil
	}
	
	logger.Debug("Repository validation successful, status has %d entries", len(status))
	return true, nil
}

func copyRemoteContentToLocalContent(remoteDirectory string, localDirectory string) error {
	// check if the destination directory exists
	// this should already be checked during the program's setup but will be
	// performed again to be safe
	exists, err := files.CheckIfDirectoryExists(localDirectory)
	if err != nil {
		logger.Error("Failed to check if local directory %s exists: %v", localDirectory, err)
		return err
	}
	if !exists {
		logger.Error("Local directory %s does not exist", localDirectory)
		return fmt.Errorf("local directory %s does not exist", localDirectory)
	}

	// copy the contents from the remote directory to the local directory
	logger.Debug("Copying contents from remote directory %s to local directory %s", remoteDirectory, localDirectory)
	completeRemoteDirectory := filepath.Join(constants.GitRemoteContentDirectory, remoteDirectory)
	err = files.RecursivelyCopyDirectory(completeRemoteDirectory, localDirectory)
	if err != nil {
		logger.Error("Failed to copy contents from remote directory %s to local directory %s: %v", remoteDirectory, localDirectory, err)
		return err
	}

	logger.Debug("Successfully copied the contents from the remote directory %s to the local directory %s", remoteDirectory, localDirectory)

	return nil
}

func HandleGitRemoteContent(serverConfig *config.ServerConfig) error {
	if serverConfig.GitRemoteContentURL == "" {
		logger.Debug("No git remote content URL provided, using local content path")
		return nil
	}

	exists, err := files.CheckIfDirectoryExists(constants.GitRemoteContentDirectory)
	if err != nil {
		logger.Error("Failed to check if git remote content directory exists: %v", err)
		return err
	}

	//  create the directory
	var directory string
	if !exists {
		logger.Debug("Git remote content directory does not exist, creating it")
		directory, err = createGitRemoteContentDirectory()
		if err != nil {
			return err
		}
	} else {
		logger.Debug("Git remote content directory exists, using it")
		directory = constants.GitRemoteContentDirectory
	}

	// pull down or clone the remote content
	// check if the git remote content directory is configured to be a git repository
	isGitRepository, err := isGitRemoteContentDirectoryAGitRepository()
	if err != nil {
		return err
	}
	// if it is not a git repository, clone the remote content
	if !isGitRepository {
		logger.Debug("Git remote content directory is not a valid git repository, preparing for fresh clone")
		
		// Clean up any existing corrupted or partial data
		// This is especially important in Docker containers where restarts might leave partial data
		logger.Debug("Cleaning up directory before cloning")
		if removeErr := os.RemoveAll(directory); removeErr != nil {
			logger.Error("Failed to clean up directory at %s: %v", directory, removeErr)
			return fmt.Errorf("failed to clean up directory: %w", removeErr)
		}
		
		// Recreate the directory
		if mkdirErr := os.MkdirAll(directory, 0750); mkdirErr != nil {
			logger.Error("Failed to recreate git remote content directory: %v", mkdirErr)
			return fmt.Errorf("failed to recreate directory: %w", mkdirErr)
		}
		
		logger.Debug("Cloning git remote content")
		err = fetchGitRemoteContent(serverConfig.GitRemoteContentURL, directory, serverConfig.GitRemoteContentBranch)
		if err != nil {
			return err
		}
	} else {
		logger.Debug("Git remote content directory is a git repository, pulling down the latest changes")
		err = pullLatestGitRemoteContent(serverConfig.GitRemoteContentBranch, directory)
		if err != nil {
			// If the repository is out of sync (e.g., "object not found" errors from shallow clones
			// after force pushes), delete it and re-clone fresh
			if strings.Contains(err.Error(), ErrOutOfSync) {
				logger.Info("Repository is out of sync, deleting and re-cloning fresh")
				// Delete the corrupted repository
				if removeErr := os.RemoveAll(directory); removeErr != nil {
					logger.Error("Failed to remove corrupted git repository at %s: %v", directory, removeErr)
					return fmt.Errorf("failed to remove corrupted repository: %w", removeErr)
				}
				// Recreate the directory
				if mkdirErr := os.MkdirAll(directory, 0750); mkdirErr != nil {
					logger.Error("Failed to recreate git remote content directory: %v", mkdirErr)
					return fmt.Errorf("failed to recreate directory: %w", mkdirErr)
				}
				// Re-clone the repository
				logger.Info("Re-cloning repository from %s", serverConfig.GitRemoteContentURL)
				err = fetchGitRemoteContent(serverConfig.GitRemoteContentURL, directory, serverConfig.GitRemoteContentBranch)
				if err != nil {
					return fmt.Errorf("failed to re-clone repository: %w", err)
				}
				logger.Info("Successfully re-cloned repository")
			} else {
				return err
			}
		}
	}

	err = syncGitRemoteDirectories(serverConfig)
	if err != nil {
		logger.Error("Failed to sync git remote directories: %v", err)
		return err
	}

	return nil
}

type directorySync struct {
	hasDirectory bool
	remotePath   string
	localPath    string
	name         string
}

func syncGitRemoteDirectories(serverConfig *config.ServerConfig) error {
	directories := []directorySync{
		{
			hasDirectory: serverConfig.HasGitRemoteContentDirectory(),
			remotePath:   serverConfig.GitRemoteContentDirectory,
			localPath:    serverConfig.ContentPath,
			name:         "content",
		},
		{
			hasDirectory: serverConfig.HasGitRemoteAssetsDirectory(),
			remotePath:   serverConfig.GitRemoteContentAssetsDirectory,
			localPath:    serverConfig.AssetsPath,
			name:         "assets",
		},
		{
			hasDirectory: serverConfig.HasGitRemoteUserStaticDirectory(),
			remotePath:   serverConfig.GitRemoteContentUserStaticDirectory,
			localPath:    serverConfig.UserStaticPath,
			name:         "user static",
		},
	}

	for _, dir := range directories {
		if !dir.hasDirectory {
			logger.Debug("No git remote %s directory provided, using local %s path", dir.name, dir.name)
			continue
		}

		if err := syncDirectory(dir.remotePath, dir.localPath, dir.name); err != nil {
			return fmt.Errorf("failed to sync %s directory: %w", dir.name, err)
		}
	}

	return nil
}

func syncDirectory(remotePath, localPath, name string) error {
	logger.Debug("Copying %s from remote directory %s to local directory %s", name, remotePath, localPath)

	// Delete existing contents
	if err := files.DeleteDirectoryContents(localPath); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", localPath, err)
	}
	logger.Debug("Successfully deleted the contents of the local directory %s", localPath)

	// Create local directory
	if err := os.MkdirAll(localPath, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", localPath, err)
	}
	logger.Debug("Successfully created the local %s directory at %s", name, localPath)

	// Copy contents
	if err := copyRemoteContentToLocalContent(remotePath, localPath); err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %w", remotePath, localPath, err)
	}
	logger.Debug("Successfully copied the contents from the remote directory %s to the local directory %s", remotePath, localPath)

	return nil
}
