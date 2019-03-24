// Package gitservice handles Git repositories on disk.
package gitservice

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"strings"
)

// GitRepository represents a repository present on disk.
type GitRepository struct {
	*git.Repository
	Directory string
}

// Checkout fetches state from origin and then updates the working tree to the given commit.
func (repo *GitRepository) Checkout(commit string) error {
	if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin"}); err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	return wt.Checkout(&git.CheckoutOptions{Force: true, Hash: plumbing.NewHash(commit)})
}

// RepoStore stores and maintains git-repos available on disk.
type RepoStore struct {
	repos     map[string]*GitRepository
	directory string
}

// NewRepoStore creates a new RepoStore that will use the given directory as its storage location.
// Returns a pointer to the new store or an error if the given directory don't exists and couldn't be created.
func NewRepoStore(directory string) (*RepoStore, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, os.ModePerm); err != nil {
			return nil, err
		}
	}
	return &RepoStore{
		repos:     map[string]*GitRepository{},
		directory: directory,
	}, nil
}

// GetRepo returns the repository with the given name.
// If the repository does not exist locally a copy will be downloaded and stored on disk.
// Returns an error if the repository didn't exist and couldn't be downloaded.
func (rs *RepoStore) GetRepo(repoName string) (*GitRepository, error) {
	repository, ok := rs.repos[repoName]
	if !ok {
		var err error
		repository, err = createRepo(rs.directory, repoName)
		if err != nil {
			return nil, err
		}
	}
	return repository, nil
}

func createRepo(parentDirectory string, repoID string) (*GitRepository, error) {
	split := strings.Split(repoID, "/")
	repoName := split[len(split)-1]
	directory := parentDirectory + "/" + repoName
	repo, err := git.PlainOpen(directory)
	if err != nil {
		repo, err = git.PlainClone(directory, false, &git.CloneOptions{
			URL: "https://github.com/" + repoID,
		})
		if err != nil {
			return nil, err
		}
	}
	return &GitRepository{
		Repository: repo,
		Directory:  directory,
	}, nil
}
