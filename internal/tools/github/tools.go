package github

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// FetchPRArgs represents the arguments for fetch_pull_request
type FetchPRArgs struct {
	PRNumber int `json:"pr_number"`
}

// FetchPRResult represents the result of fetch_pull_request
type FetchPRResult struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"`
	Author    string   `json:"author"`
	Files     []string `json:"files"`
	Additions int      `json:"additions"`
	Deletions int      `json:"deletions"`
}

// GetDiffArgs represents the arguments for get_pull_request_diff
type GetDiffArgs struct {
	PRNumber int `json:"pr_number"`
}

// GetDiffResult represents the result of get_pull_request_diff
type GetDiffResult struct {
	Diff string `json:"diff"`
}

// CreateTools creates functiontool instances for GitHub operations
func CreateTools(githubTool *Tool) ([]tool.Tool, error) {
	// Create fetch_pull_request tool
	fetchPRTool, err := functiontool.New(
		functiontool.Config{
			Name:        "fetch_pull_request",
			Description: "Fetches a pull request by number. Returns PR details including title, body, description, and changed files.",
		},
		func(ctx tool.Context, args FetchPRArgs) (FetchPRResult, error) {
			pr, err := githubTool.GetPullRequest(ctx, args.PRNumber)
			if err != nil {
				return FetchPRResult{}, err
			}

			files, err := githubTool.GetPullRequestFiles(ctx, args.PRNumber)
			if err != nil {
				return FetchPRResult{}, err
			}

			fileNames := make([]string, len(files))
			for i, f := range files {
				fileNames[i] = *f.Filename
			}

			return FetchPRResult{
				Number:    pr.GetNumber(),
				Title:     pr.GetTitle(),
				Body:      pr.GetBody(),
				State:     pr.GetState(),
				Author:    pr.GetUser().GetLogin(),
				Files:     fileNames,
				Additions: pr.GetAdditions(),
				Deletions: pr.GetDeletions(),
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetch_pull_request tool: %w", err)
	}

	// Create get_pull_request_diff tool
	diffTool, err := functiontool.New(
		functiontool.Config{
			Name:        "get_pull_request_diff",
			Description: "Gets the full diff for a pull request showing all code changes.",
		},
		func(ctx tool.Context, args GetDiffArgs) (GetDiffResult, error) {
			files, err := githubTool.GetPullRequestFiles(ctx, args.PRNumber)
			if err != nil {
				return GetDiffResult{}, err
			}

			var diffs []string
			for _, file := range files {
				if file.Patch != nil {
					diffs = append(diffs, fmt.Sprintf("File: %s\n%s\n", *file.Filename, *file.Patch))
				}
			}

			return GetDiffResult{
				Diff: strings.Join(diffs, "\n---\n\n"),
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create get_pull_request_diff tool: %w", err)
	}

	return []tool.Tool{fetchPRTool, diffTool}, nil
}

// ListMergedPRsArgs represents the arguments for list_merged_prs
type ListMergedPRsArgs struct {
	Days int `json:"days"`
}

// ListMergedPRsResult represents the result of list_merged_prs
type ListMergedPRsResult struct {
	PRs []MergedPRInfo `json:"prs"`
}

// MergedPRInfo contains information about a merged PR
type MergedPRInfo struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	MergedAt string `json:"merged_at"`
	MergeSHA string `json:"merge_sha"`
}

// CreateCherryPickArgs represents the arguments for create_cherry_pick_pr
type CreateCherryPickArgs struct {
	PRNumber     int    `json:"pr_number"`
	TargetBranch string `json:"target_branch"`
	BaseBranch   string `json:"base_branch"`
}

// CreateCherryPickResult represents the result of create_cherry_pick_pr
type CreateCherryPickResult struct {
	PRNumber int    `json:"pr_number"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
}

// CheckConflictsArgs represents the arguments for check_cherry_pick_conflicts
type CheckConflictsArgs struct {
	PRNumber     int    `json:"pr_number"`
	TargetBranch string `json:"target_branch"`
	BaseBranch   string `json:"base_branch"`
}

// CheckConflictsResult represents the result of check_cherry_pick_conflicts
type CheckConflictsResult struct {
	HasConflicts bool     `json:"has_conflicts"`
	Details      []string `json:"details"`
	Commits      int      `json:"commits"`
}

// CreateCherryPickTools creates functiontool instances for cherry-pick operations
func CreateCherryPickTools(githubTool *Tool) ([]tool.Tool, error) {
	// Create check_cherry_pick_conflicts tool
	checkConflictsTool, err := functiontool.New(
		functiontool.Config{
			Name:        "check_cherry_pick_conflicts",
			Description: "Checks if cherry-picking a merged PR's commits to a target branch would have merge conflicts. Returns whether conflicts exist and details. This should be called BEFORE create_cherry_pick_pr to verify the cherry-pick can proceed.",
		},
		func(ctx tool.Context, args CheckConflictsArgs) (CheckConflictsResult, error) {
			if args.BaseBranch == "" {
				args.BaseBranch = "main" // Default base branch
			}

			// Get commits count first
			commits, err := githubTool.GetPullRequestCommits(ctx, args.PRNumber)
			if err != nil {
				return CheckConflictsResult{}, err
			}

			hasConflicts, details, err := githubTool.CheckCherryPickConflicts(ctx, args.PRNumber, args.TargetBranch, args.BaseBranch)
			if err != nil {
				return CheckConflictsResult{}, err
			}

			return CheckConflictsResult{
				HasConflicts: hasConflicts,
				Details:      details,
				Commits:      len(commits),
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create check_cherry_pick_conflicts tool: %w", err)
	}
	// Create list_merged_prs tool
	listMergedPRsTool, err := functiontool.New(
		functiontool.Config{
			Name:        "list_merged_prs",
			Description: "Lists pull requests that were merged within the specified number of days. Returns PR numbers, titles, authors, merge dates, and merge commit SHAs.",
		},
		func(ctx tool.Context, args ListMergedPRsArgs) (ListMergedPRsResult, error) {
			if args.Days <= 0 {
				args.Days = 7 // Default to last week
			}

			prs, err := githubTool.ListMergedPullRequests(ctx, args.Days)
			if err != nil {
				return ListMergedPRsResult{}, err
			}

			prInfos := make([]MergedPRInfo, len(prs))
			for i, pr := range prs {
				mergedAt := ""
				if pr.MergedAt != nil {
					mergedAt = pr.MergedAt.Format(time.RFC3339)
				}
				prInfos[i] = MergedPRInfo{
					Number:   pr.GetNumber(),
					Title:    pr.GetTitle(),
					Author:   pr.GetUser().GetLogin(),
					MergedAt: mergedAt,
					MergeSHA: pr.GetMergeCommitSHA(),
				}
			}

			return ListMergedPRsResult{
				PRs: prInfos,
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create list_merged_prs tool: %w", err)
	}

	// Create create_cherry_pick_pr tool
	createCherryPickTool, err := functiontool.New(
		functiontool.Config{
			Name:        "create_cherry_pick_pr",
			Description: "Creates a pull request for cherry-picking a merged PR to a target branch. Only cherry-picks the commits from the PR (not the merge commit) and checks for conflicts before creating. Will fail if conflicts are detected. Only call this after: 1) checking for conflicts with check_cherry_pick_conflicts, 2) user has explicitly confirmed. Parameters: pr_number (the merged PR number to cherry-pick), target_branch (branch to cherry-pick to), base_branch (branch to create the cherry-pick branch from, default: main).",
		},
		func(ctx tool.Context, args CreateCherryPickArgs) (CreateCherryPickResult, error) {
			if args.BaseBranch == "" {
				args.BaseBranch = "main" // Default base branch
			}

			createdPR, err := githubTool.CreateCherryPickPR(ctx, args.PRNumber, args.TargetBranch, args.BaseBranch)
			if err != nil {
				return CreateCherryPickResult{}, err
			}

			return CreateCherryPickResult{
				PRNumber: createdPR.GetNumber(),
				Title:    createdPR.GetTitle(),
				URL:      createdPR.GetHTMLURL(),
				Branch:   createdPR.GetHead().GetRef(),
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create create_cherry_pick_pr tool: %w", err)
	}

	return []tool.Tool{listMergedPRsTool, checkConflictsTool, createCherryPickTool}, nil
}
