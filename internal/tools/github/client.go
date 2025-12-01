package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// Tool provides tools for interacting with GitHub
type Tool struct {
	client *github.Client
	owner  string
	repo   string
}

// NewTool creates a new GitHub tool instance
func NewTool(token, owner, repo string) (*Tool, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Tool{
		client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

// GetPullRequest fetches a pull request by number
func (g *Tool) GetPullRequest(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR #%d: %w", prNumber, err)
	}
	return pr, nil
}

// ListPullRequests lists open pull requests
func (g *Tool) ListPullRequests(ctx context.Context, state string) ([]*github.PullRequest, error) {
	if state == "" {
		state = "open"
	}

	opts := &github.PullRequestListOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	prs, _, err := g.client.PullRequests.List(ctx, g.owner, g.repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}
	return prs, nil
}

// GetPullRequestFiles gets the files changed in a pull request
func (g *Tool) GetPullRequestFiles(ctx context.Context, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := g.client.PullRequests.ListFiles(ctx, g.owner, g.repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR files: %w", err)
	}
	return files, nil
}

// GetPullRequestCommits gets the commits in a pull request (excluding merge commit)
func (g *Tool) GetPullRequestCommits(ctx context.Context, prNumber int) ([]*github.RepositoryCommit, error) {
	commits, _, err := g.client.PullRequests.ListCommits(ctx, g.owner, g.repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR commits: %w", err)
	}

	// Filter out merge commits (commits with more than one parent)
	var prCommits []*github.RepositoryCommit
	for _, commit := range commits {
		if commit.Parents != nil && len(commit.Parents) == 1 {
			// Regular commit (not a merge commit)
			prCommits = append(prCommits, commit)
		}
	}

	return prCommits, nil
}

// CheckCherryPickConflicts checks if cherry-picking PR commits to target branch would have conflicts
// This creates a test PR from the original PR's head branch to the target branch to check mergeability
func (g *Tool) CheckCherryPickConflicts(ctx context.Context, prNumber int, targetBranch string, baseBranch string) (bool, []string, error) {
	// Get the original PR
	pr, err := g.GetPullRequest(ctx, prNumber)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get PR #%d: %w", prNumber, err)
	}

	if pr.MergedAt == nil {
		return false, nil, fmt.Errorf("PR #%d is not merged", prNumber)
	}

	// Get commits from the PR (excluding merge commit)
	commits, err := g.GetPullRequestCommits(ctx, prNumber)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get PR commits: %w", err)
	}

	if len(commits) == 0 {
		return false, nil, fmt.Errorf("PR #%d has no commits to cherry-pick", prNumber)
	}

	// Get the head branch/ref from the original PR to test merging into target branch
	headRef := pr.GetHead().GetRef()
	if headRef == "" {
		return false, nil, fmt.Errorf("PR #%d does not have a head ref", prNumber)
	}

	// Create a test PR from the original PR's head to the target branch to check for conflicts
	testPRTitle := fmt.Sprintf("[TEST] Conflict check for PR #%d cherry-pick to %s", prNumber, targetBranch)
	testPRBody := fmt.Sprintf("Testing if PR #%d commits can be cherry-picked to %s. This test PR will be closed immediately.", prNumber, targetBranch)

	testPR := &github.NewPullRequest{
		Title: &testPRTitle,
		Body:  &testPRBody,
		Head:  &headRef,
		Base:  &targetBranch,
	}

	testPRCreated, _, err := g.client.PullRequests.Create(ctx, g.owner, g.repo, testPR)
	if err != nil {
		// If head ref doesn't exist (branch was deleted after merge), we can't check conflicts
		// Return a warning but don't fail - let the user know we can't verify
		return false, []string{fmt.Sprintf("Cannot check conflicts: original PR head branch '%s' may have been deleted. Proceed with caution.", headRef)}, nil
	}

	// Wait for GitHub to calculate mergeability
	time.Sleep(3 * time.Second)

	// Get the PR to check mergeable status
	testPRUpdated, _, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, testPRCreated.GetNumber())
	if err != nil {
		// Clean up
		_, _, _ = g.client.PullRequests.Edit(ctx, g.owner, g.repo, testPRCreated.GetNumber(), &github.PullRequest{State: github.String("closed")})
		return false, nil, fmt.Errorf("failed to get test PR status: %w", err)
	}

	hasConflicts := false
	var conflictDetails []string

	if testPRUpdated.Mergeable != nil {
		if !*testPRUpdated.Mergeable {
			hasConflicts = true
			conflictDetails = append(conflictDetails, fmt.Sprintf("PR #%d commits cannot be cleanly merged into %s", prNumber, targetBranch))
			if testPRUpdated.MergeableState != nil {
				conflictDetails = append(conflictDetails, fmt.Sprintf("Mergeable state: %s", *testPRUpdated.MergeableState))
			}
		}
	} else {
		// Mergeable status is still being calculated, wait a bit more
		time.Sleep(2 * time.Second)
		testPRUpdated, _, err = g.client.PullRequests.Get(ctx, g.owner, g.repo, testPRCreated.GetNumber())
		if err == nil && testPRUpdated.Mergeable != nil {
			if !*testPRUpdated.Mergeable {
				hasConflicts = true
				conflictDetails = append(conflictDetails, fmt.Sprintf("PR #%d commits cannot be cleanly merged into %s", prNumber, targetBranch))
				if testPRUpdated.MergeableState != nil {
					conflictDetails = append(conflictDetails, fmt.Sprintf("Mergeable state: %s", *testPRUpdated.MergeableState))
				}
			}
		} else {
			// Still can't determine, assume conflicts to be safe
			hasConflicts = true
			conflictDetails = append(conflictDetails, "Unable to determine mergeability status - assuming conflicts exist")
		}
	}

	// Clean up: close the test PR
	_, _, _ = g.client.PullRequests.Edit(ctx, g.owner, g.repo, testPRCreated.GetNumber(), &github.PullRequest{State: github.String("closed")})

	return hasConflicts, conflictDetails, nil
}

// ListMergedPullRequests lists merged pull requests within a time range
func (g *Tool) ListMergedPullRequests(ctx context.Context, days int) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
		Sort:      "updated",
		Direction: "desc",
	}

	var allPRs []*github.PullRequest
	for {
		prs, resp, err := g.client.PullRequests.List(ctx, g.owner, g.repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list merged PRs: %w", err)
		}

		// Filter for merged PRs within the specified days
		cutoffDate := time.Now().AddDate(0, 0, -days)
		for _, pr := range prs {
			if pr.MergedAt != nil && pr.MergedAt.After(cutoffDate) {
				allPRs = append(allPRs, pr)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// CreateCherryPickPR creates a pull request for cherry-picking a merged PR to a target branch
// It only cherry-picks the commits from the PR (not the merge commit) and checks for conflicts first
func (g *Tool) CreateCherryPickPR(ctx context.Context, prNumber int, targetBranch string, baseBranch string) (*github.PullRequest, error) {
	// Get the original PR
	pr, err := g.GetPullRequest(ctx, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR #%d: %w", prNumber, err)
	}

	if pr.MergedAt == nil {
		return nil, fmt.Errorf("PR #%d is not merged", prNumber)
	}

	// Get commits from the PR (excluding merge commit)
	commits, err := g.GetPullRequestCommits(ctx, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR commits: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("PR #%d has no commits to cherry-pick", prNumber)
	}

	// Check for conflicts before creating the PR
	hasConflicts, conflictDetails, err := g.CheckCherryPickConflicts(ctx, prNumber, targetBranch, baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to check for conflicts: %w", err)
	}

	if hasConflicts {
		details := "Unknown conflicts"
		if len(conflictDetails) > 0 {
			details = conflictDetails[0]
		}
		return nil, fmt.Errorf("cannot cherry-pick PR #%d to %s: %s", prNumber, targetBranch, details)
	}

	// Create a new branch name for the cherry-pick
	cherryPickBranch := fmt.Sprintf("cherry-pick-%d-to-%s", prNumber, targetBranch)

	// Get the target branch SHA (we cherry-pick to target branch, not base)
	targetRef, _, err := g.client.Git.GetRef(ctx, g.owner, g.repo, "refs/heads/"+targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get target branch %s: %w", targetBranch, err)
	}

	// Create a new branch from the target branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + cherryPickBranch),
		Object: &github.GitObject{
			SHA: targetRef.Object.SHA,
		},
	}
	_, _, err = g.client.Git.CreateRef(ctx, g.owner, g.repo, newRef)
	if err != nil {
		// Branch might already exist, try to get it
		existingRef, _, getErr := g.client.Git.GetRef(ctx, g.owner, g.repo, "refs/heads/"+cherryPickBranch)
		if getErr != nil {
			return nil, fmt.Errorf("failed to create branch %s: %w", cherryPickBranch, err)
		}
		newRef = existingRef
	}

	// Build list of commit SHAs for the PR body
	var commitSHAs []string
	for _, commit := range commits {
		commitSHAs = append(commitSHAs, commit.GetSHA())
	}

	// Create PR title and body
	title := fmt.Sprintf("[%s] %s", cherryPickBranch, pr.GetTitle())
	body := fmt.Sprintf("This is a cherry-pick of PR #%d to %s.\n\n"+
		"Original PR: #%d\n"+
		"Original Author: @%s\n"+
		"Original Merge Date: %s\n"+
		"Commits being cherry-picked: %d\n\n"+
		"**Note**: This PR contains only the commits from the original PR (excluding merge commit).\n"+
		"To cherry-pick manually:\n"+
		"```\n"+
		"git checkout %s\n"+
		"git cherry-pick %s\n"+
		"```\n",
		prNumber, targetBranch, prNumber, pr.GetUser().GetLogin(), pr.MergedAt.Format("2006-01-02"), len(commits), cherryPickBranch, strings.Join(commitSHAs, " "))

	// Create the pull request
	newPR := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &cherryPickBranch,
		Base:  &targetBranch,
	}

	createdPR, _, err := g.client.PullRequests.Create(ctx, g.owner, g.repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create cherry-pick PR: %w", err)
	}

	return createdPR, nil
}

// LoadFromEnv creates a GitHub tool from environment variables
func LoadFromEnv() (*Tool, error) {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_ORG")
	repo := os.Getenv("GITHUB_REPO")

	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_ORG environment variable is required")
	}
	if repo == "" {
		return nil, fmt.Errorf("GITHUB_REPO environment variable is required")
	}

	return NewTool(token, owner, repo)
}
