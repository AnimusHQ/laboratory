package domain

import "time"

// CodeRef identifies the exact source used for execution.
type CodeRef struct {
	RepoURL   string `json:"repoUrl"`
	CommitSHA string `json:"commitSha"`
	Path      string `json:"path,omitempty"`
	SCMType   string `json:"scmType,omitempty"`
}

// EnvironmentDefinition describes a logical execution environment.
type EnvironmentDefinition struct {
	ID              string    `json:"environmentDefinitionId"`
	ProjectID       string    `json:"projectId"`
	Name            string    `json:"name"`
	Version         int       `json:"version"`
	Description     string    `json:"description,omitempty"`
	BaseImageRef    string    `json:"baseImageRef"`
	Metadata        Metadata  `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	CreatedBy       string    `json:"createdBy,omitempty"`
	IntegritySHA256 string    `json:"integritySha256"`
}

// EnvLock captures the immutable execution environment bindings.
type EnvLock struct {
	LockID              string            `json:"lockId,omitempty"`
	ImageDigests        map[string]string `json:"imageDigests"`
	DependencyChecksums map[string]string `json:"dependencyChecksums,omitempty"`
	EnvTemplateID       string            `json:"envTemplateId,omitempty"`
	EnvHash             string            `json:"envHash"`
	SBOMRef             string            `json:"sbomRef,omitempty"`
}

// PolicySnapshot captures the governance and policy context at run creation.
type PolicySnapshot struct {
	SnapshotVersion string                  `json:"snapshotVersion"`
	CapturedAt      time.Time               `json:"capturedAt"`
	CapturedBy      string                  `json:"capturedBy,omitempty"`
	RBAC            PolicySnapshotRBAC      `json:"rbac"`
	Retention       PolicySnapshotRetention `json:"retention,omitempty"`
	Network         PolicySnapshotNetwork   `json:"network,omitempty"`
	Templates       PolicySnapshotTemplates `json:"templates,omitempty"`
	Policies        []PolicySnapshotPolicy  `json:"policies"`
	SnapshotSHA256  string                  `json:"snapshotSha256"`
}

type PolicySnapshotRBAC struct {
	Subject   string   `json:"subject"`
	Roles     []string `json:"roles"`
	ProjectID string   `json:"projectId"`
}

type PolicySnapshotRetention struct {
	Mode            string `json:"mode"`
	PolicyID        string `json:"policyId,omitempty"`
	PolicyVersionID string `json:"policyVersionId,omitempty"`
	PolicySHA256    string `json:"policySha256,omitempty"`
	LegalHold       bool   `json:"legalHold,omitempty"`
}

type PolicySnapshotNetwork struct {
	Mode      string   `json:"mode"`
	Allowlist []string `json:"allowlist,omitempty"`
	Denylist  []string `json:"denylist,omitempty"`
}

type PolicySnapshotTemplates struct {
	Mode               string   `json:"mode"`
	AllowedTemplateIDs []string `json:"allowedTemplateIds,omitempty"`
}

type PolicySnapshotPolicy struct {
	PolicyID        string `json:"policyId"`
	PolicyName      string `json:"policyName,omitempty"`
	PolicyVersionID string `json:"policyVersionId"`
	PolicyVersion   int    `json:"policyVersion,omitempty"`
	PolicySHA256    string `json:"policySha256"`
	Status          string `json:"status"`
}
