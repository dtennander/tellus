// Package for writing Tellus comments and checks to Github.com.
package ghclient

import (
	"context"
	"github.com/google/go-github/v24/github"
	"log"
	"time"
)


// Interface mapped against IssuesService in "github.com/google/go-github/v24/github"
type issueService interface {
	CreateComment(
		ctx context.Context,
		owner string,
		repo string,
		prNumber int,
		comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

// Interface mapped against ChecksService in "github.com/google/go-github/v24/github"
type checksService interface {
	CreateCheckRun(
		ctx context.Context,
		Name string,
		repo string,
		options github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
}

// Client handling the information sent to Github.
type Client struct {
	issuesService issueService
	checksService checksService
}

// Creates a new instance of the Client.
func NewClient(service issueService, service2 checksService) *Client {
	return &Client{
		issuesService:service,
		checksService:service2,
	}
}

// Create a Tellus comment presenting the output
// on pull request prNumber on repository repo belonging to the given owner.
// The comment will have the format:
//    Tellus ran `terraform plan` on this PR and got:
//    ```
//    <output>
//    ```
func (c *Client) CreateComment(output string, owner string, repo string, prNumber int) error {
	result := "Tellus ran `terraform plan` on this PR and got:\n```\n" + output + "\n```"
	_, _, err := c.issuesService.CreateComment(
		context.Background(),
		owner,
		repo,
		prNumber,
		&github.IssueComment{Body: &result})
	return err
}

// Creates a Tellus status on the commit on the repository repo with the given owner.
// The status will ether be a success or a failure given the boolean flag success.
// The status will have the following structure:
//    Summary: Tellus have run terraform <tfCommand>
//    Title: Tellus <tfCommand>
//    Name: Tellus have run terraform <tfCommand>
func (c *Client) CreateCommitStatus(owner string, repo string, commit string, success bool, output string, tfCommand string) error {
	completed := "completed"
	var conclusion string
	if success {
		conclusion = "success"
	} else {
		conclusion = "failure"
	}
	log.Printf("Reporting back %s status on commit %s", conclusion, commit)
	summary := "Tellus have run terraform " + tfCommand
	title := "Tellus: " + tfCommand
	_, _, err := c.checksService.CreateCheckRun(
		context.Background(),
		owner,
		repo,
		github.CreateCheckRunOptions{
			Name:       "Tellus have run terraform " + tfCommand,
			HeadSHA:    commit,
			Status:     &completed,
			Conclusion: &conclusion,
			CompletedAt: &github.Timestamp{
				Time: time.Now(),
			},
			Output: &github.CheckRunOutput{
				Title:   &title,
				Text:    &output,
				Summary: &summary,
			},
		})
	return err
}