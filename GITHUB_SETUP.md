# GitHub Tool Setup Guide

## Required Environment Variables

Set these in your `env` file:

### 1. GITHUB_TOKEN
- **Type**: String
- **Required**: Yes
- **Description**: GitHub Personal Access Token (PAT)
- **How to get**:
  1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
  2. Click "Generate new token (classic)"
  3. Select scopes: `repo` (for private repos) or `public_repo` (for public repos)
  4. Copy the token and add it to your `env` file

**Example**: `GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### 2. GITHUB_ORG
- **Type**: String
- **Required**: Yes
- **Description**: GitHub organization name or username that owns the repository
- **Example**: 
  - For `https://github.com/openshift/kubernetes` → `GITHUB_ORG=openshift`
  - For `https://github.com/ramishra/my-repo` → `GITHUB_ORG=ramishra`

### 3. GITHUB_REPO
- **Type**: String
- **Required**: Yes
- **Description**: Repository name (without the organization/username)
- **Example**:
  - For `https://github.com/openshift/kubernetes` → `GITHUB_REPO=kubernetes`
  - For `https://github.com/ramishra/my-repo` → `GITHUB_REPO=my-repo`

## Available GitHub Tools/Functions

The GitHub tool provides two functions that agents can use:

### 1. `fetch_pull_request`
- **Purpose**: Fetches pull request details
- **Parameters**:
  - `pr_number` (integer, required): The pull request number
- **Returns**: PR details including:
  - Title, body, description
  - Author, creation date
  - List of changed files
  - Additions/deletions count
  - State (open/closed)

### 2. `get_pull_request_diff`
- **Purpose**: Gets the full code diff for a pull request
- **Parameters**:
  - `pr_number` (integer, required): The pull request number
- **Returns**: Complete diff showing all code changes

## Example Configuration

```bash
# GitHub Configuration
GITHUB_TOKEN=ghp_your_token_here
GITHUB_ORG=openshift
GITHUB_REPO=kubernetes
```

## Usage Flow

1. **Agent asks for PR number**: When you interact with the agent, provide a PR number
   - Example: "Review PR #123"
   - Example: "Fetch details for pull request 456"

2. **PR Fetcher Agent**: Uses `fetch_pull_request` to get PR details

3. **PR Reviewer Agent**: Uses `get_pull_request_diff` to analyze code changes

4. **Workflow**: Sequentially executes both agents to complete the review

## Token Permissions Required

Your GitHub token needs these permissions:
- **repo** (full control of private repositories) - for private repos
- **public_repo** (access public repositories) - for public repos only

For most use cases, `repo` scope is recommended.

