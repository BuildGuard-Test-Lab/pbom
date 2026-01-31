package github

import "time"

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	HeadSHA      string    `json:"head_sha"`
	HeadBranch   string    `json:"head_branch"`
	Path         string    `json:"path"`
	DisplayTitle string    `json:"display_title"`
	Event        string    `json:"event"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	WorkflowID   int64     `json:"workflow_id"`
	Actor        Actor     `json:"actor"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RunStartedAt time.Time `json:"run_started_at"`
	Repository   Repo      `json:"repository"`
}

// WorkflowRunsResponse represents the list runs API response.
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// Actor represents a GitHub user.
type Actor struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// Repo represents a GitHub repository (minimal fields).
type Repo struct {
	ID       int64 `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    Owner  `json:"owner"`
}

// Owner represents a repository owner.
type Owner struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// JobsResponse represents the list jobs API response.
type JobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}

// Job represents a single job within a workflow run.
type Job struct {
	ID              int64     `json:"id"`
	RunID           int64     `json:"run_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Conclusion      string    `json:"conclusion"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	Labels          []string  `json:"labels"`
	RunnerName      string    `json:"runner_name"`
	RunnerID        int64     `json:"runner_id"`
	RunnerGroupName string    `json:"runner_group_name"`
	Steps           []Step    `json:"steps"`
}

// Step represents a single step within a job.
type Step struct {
	Name        string    `json:"name"`
	Number      int       `json:"number"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// ArtifactsResponse represents the list artifacts API response.
type ArtifactsResponse struct {
	TotalCount int        `json:"total_count"`
	Artifacts  []Artifact `json:"artifacts"`
}

// Artifact represents a workflow run artifact.
type Artifact struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	SizeInBytes        int64     `json:"size_in_bytes"`
	ArchiveDownloadURL string    `json:"archive_download_url"`
	Digest             string    `json:"digest"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
}

// FileContent represents a file fetched via the Contents API.
type FileContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	Path     string `json:"path"`
	SHA      string `json:"sha"`
}
