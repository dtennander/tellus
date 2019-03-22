package tellus

import (
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v24/github"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"tellus/tellus/ghclient"
	"tellus/tellus/gitservice"
	"tellus/tellus/terraform"
)

type Client struct {
	repositories  *gitservice.RepoStore
	output 		  *ghclient.Client
}

func NewClient(keyFile string, repoDirectory string, integrationId int, installationId int) (*Client, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, integrationId, installationId, keyFile)
	if err != nil {
		return nil, err
	}
	client := github.NewClient(&http.Client{Transport: itr})
	store, err := gitservice.NewRepoStore(repoDirectory)
	if err != nil {
		return nil, err
	}
	return &Client{
		repositories: store,
		output: ghclient.NewClient(client.Issues, client.Checks),
	}, nil
}

// Handle a new Pull Request.
// This will checkout the code, run terraform and send PR comment together with status check on commit.
func (c *Client) NewPR(payload *github.PullRequestEvent) error {
	repo, err := c.checkoutCode(*payload.Repo.FullName, *payload.PullRequest.Head.SHA)
	tfDirectory, err := getTfDirs(repo.Directory)
	if err != nil || tfDirectory == "" {
		return err
	}
	output, ok := terraform.Plan(tfDirectory)
	owner := *payload.Repo.Owner.Login
	repoName := *payload.Repo.Name
	err = c.output.CreateCommitStatus(owner, repoName, *payload.PullRequest.Head.SHA, ok, output, "plan")
	if err != nil {
		return err
	}
	err = c.output.CreateComment(output, owner, repoName, *payload.Number, ok)
	if err != nil {
		return err
	}
	return nil
}

func getTfDirs(baseDir string) (string ,error){
	file, err := os.Open(baseDir + "/.tellus")
	if err != nil {
		return "", err
	}
	var config Configuration
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return "", err
	}
	return config.TerraformDirectory, nil
}


func (c *Client) checkoutCode(repoName string, commit string) (*gitservice.GitRepository, error) {
	log.Printf("Checking our commit %s on %s", commit, repoName)
	log.Println(c)
	repo, err := c.repositories.GetRepo(repoName)
	if err != nil {
		return nil, err
	}
	err = repo.Checkout(commit)
	return repo, err
}

func (c *Client) NewPush(payload *github.PushEvent) error {
	fullName := *payload.Repo.FullName
	commit := *payload.HeadCommit.ID
	if *payload.Ref != "refs/heads/master" {
		log.Printf("Ignoring push to %s", *payload.Ref)
		return nil
	}
	repo, err := c.checkoutCode(fullName, commit)
	output, ok := terraform.Apply(repo.Directory)
	owner := *payload.Repo.Owner.Name
	repoName := *payload.Repo.Name
	err = c.output.CreateCommitStatus(owner, repoName, commit, ok, output, "apply")
	if err != nil {
		return err
	}
	return nil
}

