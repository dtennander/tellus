// Package tellus is a Bot that integrates terraform CI/CD into any github repository.
// Tellus assumes that it is run in an environment that is
// authenticated and authorized to execute the terraform commands.
// It also needs git and terraform present on the running environment.
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

// Client is the main API of the Tellus bot.
type Client struct {
	repositories *gitservice.RepoStore
	output       *ghclient.Client
}

// NewClient creates a new Tellus client.
//
// keyFile should be the private key given by Github to authenticate the app,
// repoDirectory is the directory in which all repositories downloaded will be stored.
// integrationId and installationID are ids given by Github to each App <-> Installation.
func NewClient(keyFile string, repoDirectory string, integrationID int, installationID int) (*Client, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, integrationID, installationID, keyFile)
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
		output:       ghclient.NewClient(client.Issues, client.Checks),
	}, nil
}

// NewPR handles a new Pull Request.
// This will checkout the code, run terraform and send PR comment together with status check on commit.
func (c *Client) NewPR(payload *github.PullRequestEvent) error {
	repo, err := c.checkoutCode(*payload.Repo.FullName, *payload.PullRequest.Head.SHA)
	config, err := getRepoConfig(repo.Directory)
	tfDir := repo.Directory + "/" + config.TerraformDirectory
	log.Printf("found TF directory: %s", tfDir)
	output, ok := terraform.Plan(tfDir)
	owner := *payload.Repo.Owner.Login
	repoName := *payload.Repo.Name
	err = c.output.CreateCommitStatus(owner, repoName, *payload.PullRequest.Head.SHA, ok, output, "plan")
	if err != nil {
		return err
	}
	return c.output.CreateComment(output, owner, repoName, *payload.Number)
}

func getRepoConfig(baseDir string) (*Configuration, error) {
	file, err := os.Open(baseDir + "/.tellus")
	if err != nil {
		return nil, err
	}
	config := NewDefaultConfig()
	err = yaml.NewDecoder(file).Decode(config)
	return config, err
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

// NewPush handles a new push event from Github.
// If the pushed branch is master
// this will checkout the code, run terraform apply and send commit status bach with the result.
func (c *Client) NewPush(payload *github.PushEvent) error {
	fullName := *payload.Repo.FullName
	commit := *payload.HeadCommit.ID
	repo, err := c.checkoutCode(fullName, commit)
	config, err := getRepoConfig(repo.Directory)
	if err != nil {
		return err
	}
	if *payload.Ref != "refs/heads/"+config.Branch {
		log.Printf("Ignoring push to %s", *payload.Ref)
		return nil
	}
	tfDir := repo.Directory + "/" + config.TerraformDirectory
	log.Printf("found TF directory: %s", tfDir)
	output, ok := terraform.Apply(tfDir)
	owner := *payload.Repo.Owner.Name
	repoName := *payload.Repo.Name
	return c.output.CreateCommitStatus(owner, repoName, commit, ok, output, "apply")
}

// CheckRunEvent handles check_run events and reruns checks if the action is "rerequested"
func (c *Client) CheckRunEvent(event *github.CheckRunEvent) error {
	if *event.Action != "rerequested" {
		log.Printf("Ignoring check_run event action: %s", *event.Action)
		return nil
	}
	fullName := *event.Repo.FullName
	commit := *event.CheckRun.HeadSHA
	repo, err := c.checkoutCode(fullName, commit)
	config, err := getRepoConfig(repo.Directory)
	if err != nil {
		return err
	}
	tfDir := repo.Directory + "/" + config.TerraformDirectory
	var result, command string
	var ok bool
	if *event.CheckRun.CheckSuite.HeadBranch == config.Branch {
		command = "apply"
		result, ok = terraform.Apply(tfDir)
	} else {
		command = "plan"
		result, ok = terraform.Plan(tfDir)
	}
	repoOwner := *event.Repo.Owner.Login //.Name does not exist...
	repoName := *event.Repo.Name
	return c.output.CreateCommitStatus(repoOwner, repoName, commit, ok, result, command)
}
