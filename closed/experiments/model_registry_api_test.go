package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
)

func TestCreateModelVersionIdempotent(t *testing.T) {
	modelStore := newStubModelStore()
	modelStore.models["model-1"] = domain.Model{
		ID:        "model-1",
		ProjectID: "proj-1",
		Name:      "model",
		Version:   "v1",
		Status:    domain.ModelStatusDraft,
	}
	versionStore := newStubModelVersionStore()
	transitionStore := &stubModelVersionTransitionStore{}
	audit := &captureAudit{}
	lineage := &captureLineage{}
	runBindings := stubRunBindingsStore{
		codeRef:   domain.CodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
		envLock:   domain.EnvLock{LockID: "lock-1"},
		policySHA: "policy-sha",
	}

	api := &experimentsAPI{
		modelStoreOverride:             modelStore,
		modelVersionStoreOverride:      versionStore,
		modelVersionTransitionOverride: transitionStore,
		modelRunBindingsOverride:       runBindings,
		modelAuditOverride:             audit,
		modelLineageOverride:           lineage,
	}

	body := `{"version":"v1","runId":"run-1","artifactIds":["a1"],"datasetVersionIds":["d1"]}`
	req := httptest.NewRequest(http.MethodPost, "/projects/proj-1/models/model-1/versions", bytes.NewBufferString(body))
	req.Header.Set("Idempotency-Key", "idem-1")
	req.SetPathValue("project_id", "proj-1")
	req.SetPathValue("model_id", "model-1")
	req = req.WithContext(auth.ContextWithIdentity(req.Context(), auth.Identity{Subject: "user-1"}))
	resp := httptest.NewRecorder()
	api.handleCreateModelVersion(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", resp.Code)
	}
	var first modelVersionCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&first); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !first.Created {
		t.Fatalf("expected created=true")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/projects/proj-1/models/model-1/versions", bytes.NewBufferString(body))
	req2.Header.Set("Idempotency-Key", "idem-1")
	req2.SetPathValue("project_id", "proj-1")
	req2.SetPathValue("model_id", "model-1")
	req2 = req2.WithContext(auth.ContextWithIdentity(req2.Context(), auth.Identity{Subject: "user-1"}))
	resp2 := httptest.NewRecorder()
	api.handleCreateModelVersion(resp2, req2)

	if resp2.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", resp2.Code)
	}
	var second modelVersionCreateResponse
	if err := json.NewDecoder(resp2.Body).Decode(&second); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if second.Created {
		t.Fatalf("expected created=false on idempotent call")
	}
	if first.ModelVersion.ID != second.ModelVersion.ID {
		t.Fatalf("expected same model version id")
	}
	if audit.count(auditModelVersionCreated) != 1 {
		t.Fatalf("expected single created audit")
	}
}

func TestModelVersionTransitionInvalid(t *testing.T) {
	versionStore := newStubModelVersionStore()
	versionStore.versions["ver-1"] = domain.ModelVersion{
		ID:        "ver-1",
		ProjectID: "proj-1",
		ModelID:   "model-1",
		Status:    domain.ModelStatusApproved,
	}
	transitionStore := &stubModelVersionTransitionStore{}
	audit := &captureAudit{}

	api := &experimentsAPI{
		modelVersionStoreOverride:      versionStore,
		modelVersionTransitionOverride: transitionStore,
		modelAuditOverride:             audit,
	}

	req := httptest.NewRequest(http.MethodPost, "/projects/proj-1/model-versions/ver-1:validate", nil)
	req.SetPathValue("project_id", "proj-1")
	req.SetPathValue("model_version_id", "ver-1")
	req = req.WithContext(auth.ContextWithIdentity(req.Context(), auth.Identity{Subject: "user-1"}))
	resp := httptest.NewRecorder()
	api.handleValidateModelVersion(resp, req)

	if resp.Code != http.StatusConflict {
		t.Fatalf("status=%d want 409", resp.Code)
	}
	if versionStore.updateCalls != 0 {
		t.Fatalf("expected no update calls")
	}
}

func TestModelVersionApproveAudited(t *testing.T) {
	versionStore := newStubModelVersionStore()
	versionStore.versions["ver-1"] = domain.ModelVersion{
		ID:        "ver-1",
		ProjectID: "proj-1",
		ModelID:   "model-1",
		Status:    domain.ModelStatusValidated,
		CreatedAt: time.Now().UTC(),
	}
	transitionStore := &stubModelVersionTransitionStore{}
	audit := &captureAudit{}

	api := &experimentsAPI{
		modelVersionStoreOverride:      versionStore,
		modelVersionTransitionOverride: transitionStore,
		modelAuditOverride:             audit,
	}

	req := httptest.NewRequest(http.MethodPost, "/projects/proj-1/model-versions/ver-1:approve", nil)
	req.SetPathValue("project_id", "proj-1")
	req.SetPathValue("model_version_id", "ver-1")
	req = req.WithContext(auth.ContextWithIdentity(req.Context(), auth.Identity{Subject: "user-1"}))
	resp := httptest.NewRecorder()
	api.handleApproveModelVersion(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", resp.Code)
	}
	if versionStore.updateCalls != 1 {
		t.Fatalf("expected status update")
	}
	if audit.count(auditModelVersionApproved) != 1 {
		t.Fatalf("expected approved audit")
	}
	if len(transitionStore.transitions) != 1 {
		t.Fatalf("expected transition record")
	}
}
