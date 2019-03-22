package gitservice

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"strings"
)

type GitRepository struct {
	*git.Repository
	Directory string
}

// Fetches state from upstream and then updates the working tree to the given commit.
func (repo *GitRepository) Checkout(commit string) error {
	if err := repo.Fetch(&git.FetchOptions{RemoteName:"origin"}); err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.Checkout(&git.CheckoutOptions{Force:true, Hash:plumbing.NewHash(commit)}); err != nil {
		return err
	}
	return nil
}

type RepoStore struct {
	repos map[string]*GitRepository
	directory string
}

func NewRepoStore(directory string) (*RepoStore, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.Mkdir(directory, os.ModePerm); err != nil {
			return nil, err
		}
	}

	return &RepoStore{
		repos: map[string]*GitRepository{},
		directory: directory,
	}, nil
}

func (rs *RepoStore) GetRepo(repoId string) (*GitRepository, error) {
	repository, ok  := rs.repos[repoId]
	if !ok {
		var err error
		repository, err  = createRepo(rs.directory, repoId)
		if err != nil {
			return nil, err
		}
	}
	return repository, nil
}

func createRepo(parentDirectory string, repoId string) (*GitRepository, error) {
	split := strings.Split(repoId, "/")
	repoName := split[len(split)-1]
	directory := parentDirectory + "/" + repoName
	repo, err := git.PlainOpen(directory)
	if err != nil {
		repo, err = git.PlainClone(directory, false, &git.CloneOptions{
			URL: "https://github.com/" + repoId,
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


