package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

type RunBindingsStore struct {
	db DB
}

const (
	insertRunCodeRefQuery = `INSERT INTO run_code_refs (
			run_id,
			project_id,
			repo_url,
			commit_sha,
			path,
			scm_type,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (run_id) DO NOTHING`
	insertRunEnvLockQuery = `INSERT INTO run_environment_locks (
			lock_id,
			run_id,
			project_id,
			env_hash,
			env_template_id,
			image_digests,
			dependency_checksums,
			sbom_ref,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (run_id) DO NOTHING`
	insertRunPolicySnapshotQuery = `INSERT INTO run_policy_snapshots (
			snapshot_id,
			run_id,
			project_id,
			snapshot,
			snapshot_sha256,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (run_id) DO NOTHING`
	selectRunCodeRefQuery = `SELECT repo_url, commit_sha, path, scm_type
		 FROM run_code_refs
		 WHERE project_id = $1 AND run_id = $2`
	selectRunEnvLockQuery = `SELECT lock_id, env_hash, env_template_id, image_digests, dependency_checksums, sbom_ref
		 FROM run_environment_locks
		 WHERE project_id = $1 AND run_id = $2`
	selectRunPolicySnapshotQuery = `SELECT snapshot
		 FROM run_policy_snapshots
		 WHERE project_id = $1 AND run_id = $2`
	selectRunPolicySnapshotSHAQuery = `SELECT snapshot_sha256
		 FROM run_policy_snapshots
		 WHERE project_id = $1 AND run_id = $2`
	selectEnvDefinitionExistsQuery = `SELECT 1 FROM environment_definitions WHERE project_id = $1 AND environment_definition_id = $2`
)

func NewRunBindingsStore(db DB) *RunBindingsStore {
	if db == nil {
		return nil
	}
	return &RunBindingsStore{db: db}
}

func (s *RunBindingsStore) InsertCodeRef(ctx context.Context, runID, projectID string, ref domain.CodeRef, createdAt time.Time, createdBy, integritySHA string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("run bindings store not initialized")
	}
	runID = strings.TrimSpace(runID)
	projectID = strings.TrimSpace(projectID)
	if runID == "" {
		return fmt.Errorf("run id is required")
	}
	if projectID == "" {
		return fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(ref.RepoURL) == "" {
		return fmt.Errorf("repo url is required")
	}
	if strings.TrimSpace(ref.CommitSHA) == "" {
		return fmt.Errorf("commit sha is required")
	}
	if strings.TrimSpace(createdBy) == "" {
		return fmt.Errorf("created by is required")
	}
	if err := requireIntegrity(integritySHA); err != nil {
		return err
	}
	createdAt = normalizeTime(createdAt)
	_, err := s.db.ExecContext(
		ctx,
		insertRunCodeRefQuery,
		runID,
		projectID,
		strings.TrimSpace(ref.RepoURL),
		strings.TrimSpace(ref.CommitSHA),
		nullIfEmpty(ref.Path),
		nullIfEmpty(ref.SCMType),
		createdAt,
		strings.TrimSpace(createdBy),
		strings.TrimSpace(integritySHA),
	)
	if err != nil {
		return fmt.Errorf("insert run code ref: %w", err)
	}
	return nil
}

func (s *RunBindingsStore) InsertEnvLock(ctx context.Context, runID, projectID string, lock domain.EnvLock, createdAt time.Time, createdBy, integritySHA string) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("run bindings store not initialized")
	}
	runID = strings.TrimSpace(runID)
	projectID = strings.TrimSpace(projectID)
	if runID == "" {
		return "", fmt.Errorf("run id is required")
	}
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(lock.EnvHash) == "" {
		return "", fmt.Errorf("env hash is required")
	}
	if strings.TrimSpace(createdBy) == "" {
		return "", fmt.Errorf("created by is required")
	}
	if err := requireIntegrity(integritySHA); err != nil {
		return "", err
	}
	createdAt = normalizeTime(createdAt)
	imageDigestsJSON, err := json.Marshal(lock.ImageDigests)
	if err != nil {
		return "", fmt.Errorf("encode image digests: %w", err)
	}
	dependencyJSON, err := json.Marshal(lock.DependencyChecksums)
	if err != nil {
		return "", fmt.Errorf("encode dependency checksums: %w", err)
	}
	lockID := strings.TrimSpace(lock.LockID)
	if lockID == "" {
		return "", fmt.Errorf("lock id is required")
	}
	_, err = s.db.ExecContext(
		ctx,
		insertRunEnvLockQuery,
		lockID,
		runID,
		projectID,
		strings.TrimSpace(lock.EnvHash),
		nullIfEmpty(lock.EnvTemplateID),
		imageDigestsJSON,
		dependencyJSON,
		nullIfEmpty(lock.SBOMRef),
		createdAt,
		strings.TrimSpace(createdBy),
		strings.TrimSpace(integritySHA),
	)
	if err != nil {
		return "", fmt.Errorf("insert run env lock: %w", err)
	}
	return lockID, nil
}

func (s *RunBindingsStore) InsertPolicySnapshot(ctx context.Context, runID, projectID string, snapshot domain.PolicySnapshot, snapshotJSON []byte, createdAt time.Time, createdBy, integritySHA string) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("run bindings store not initialized")
	}
	runID = strings.TrimSpace(runID)
	projectID = strings.TrimSpace(projectID)
	if runID == "" {
		return "", fmt.Errorf("run id is required")
	}
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(snapshot.SnapshotSHA256) == "" {
		return "", fmt.Errorf("snapshot sha256 is required")
	}
	if strings.TrimSpace(createdBy) == "" {
		return "", fmt.Errorf("created by is required")
	}
	if err := requireIntegrity(integritySHA); err != nil {
		return "", err
	}
	createdAt = normalizeTime(createdAt)
	if len(snapshotJSON) == 0 {
		return "", fmt.Errorf("snapshot json is required")
	}
	snapshotID := strings.TrimSpace(snapshot.SnapshotSHA256)
	_, err := s.db.ExecContext(
		ctx,
		insertRunPolicySnapshotQuery,
		snapshotID,
		runID,
		projectID,
		snapshotJSON,
		strings.TrimSpace(snapshot.SnapshotSHA256),
		createdAt,
		strings.TrimSpace(createdBy),
		strings.TrimSpace(integritySHA),
	)
	if err != nil {
		return "", fmt.Errorf("insert run policy snapshot: %w", err)
	}
	return snapshotID, nil
}

func (s *RunBindingsStore) GetCodeRef(ctx context.Context, projectID, runID string) (domain.CodeRef, error) {
	if s == nil || s.db == nil {
		return domain.CodeRef{}, fmt.Errorf("run bindings store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	if projectID == "" || runID == "" {
		return domain.CodeRef{}, fmt.Errorf("project id and run id are required")
	}
	var ref domain.CodeRef
	row := s.db.QueryRowContext(
		ctx,
		selectRunCodeRefQuery,
		projectID,
		runID,
	)
	var path sql.NullString
	var scm sql.NullString
	if err := row.Scan(&ref.RepoURL, &ref.CommitSHA, &path, &scm); err != nil {
		return domain.CodeRef{}, handleNotFound(err)
	}
	if path.Valid {
		ref.Path = path.String
	}
	if scm.Valid {
		ref.SCMType = scm.String
	}
	return ref, nil
}

func (s *RunBindingsStore) GetEnvLock(ctx context.Context, projectID, runID string) (domain.EnvLock, error) {
	if s == nil || s.db == nil {
		return domain.EnvLock{}, fmt.Errorf("run bindings store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	if projectID == "" || runID == "" {
		return domain.EnvLock{}, fmt.Errorf("project id and run id are required")
	}
	var (
		lock         domain.EnvLock
		imageDigests []byte
		depsJSON     []byte
		envTemplate  sql.NullString
		sbom         sql.NullString
	)
	row := s.db.QueryRowContext(
		ctx,
		selectRunEnvLockQuery,
		projectID,
		runID,
	)
	if err := row.Scan(&lock.LockID, &lock.EnvHash, &envTemplate, &imageDigests, &depsJSON, &sbom); err != nil {
		return domain.EnvLock{}, handleNotFound(err)
	}
	if envTemplate.Valid {
		lock.EnvTemplateID = envTemplate.String
	}
	if sbom.Valid {
		lock.SBOMRef = sbom.String
	}
	if len(imageDigests) > 0 {
		var digests map[string]string
		if err := json.Unmarshal(imageDigests, &digests); err != nil {
			return domain.EnvLock{}, fmt.Errorf("decode image digests: %w", err)
		}
		lock.ImageDigests = digests
	}
	if len(depsJSON) > 0 {
		var deps map[string]string
		if err := json.Unmarshal(depsJSON, &deps); err != nil {
			return domain.EnvLock{}, fmt.Errorf("decode dependency checksums: %w", err)
		}
		lock.DependencyChecksums = deps
	}
	return lock, nil
}

func (s *RunBindingsStore) GetPolicySnapshot(ctx context.Context, projectID, runID string) (domain.PolicySnapshot, error) {
	if s == nil || s.db == nil {
		return domain.PolicySnapshot{}, fmt.Errorf("run bindings store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	if projectID == "" || runID == "" {
		return domain.PolicySnapshot{}, fmt.Errorf("project id and run id are required")
	}
	var snapshotJSON []byte
	row := s.db.QueryRowContext(
		ctx,
		selectRunPolicySnapshotQuery,
		projectID,
		runID,
	)
	if err := row.Scan(&snapshotJSON); err != nil {
		return domain.PolicySnapshot{}, handleNotFound(err)
	}
	var snapshot domain.PolicySnapshot
	if err := json.Unmarshal(snapshotJSON, &snapshot); err != nil {
		return domain.PolicySnapshot{}, fmt.Errorf("decode policy snapshot: %w", err)
	}
	return snapshot, nil
}

func (s *RunBindingsStore) PolicySnapshotSHA(ctx context.Context, projectID, runID string) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("run bindings store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	runID = strings.TrimSpace(runID)
	if projectID == "" || runID == "" {
		return "", fmt.Errorf("project id and run id are required")
	}
	row := s.db.QueryRowContext(
		ctx,
		selectRunPolicySnapshotSHAQuery,
		projectID,
		runID,
	)
	var value string
	if err := row.Scan(&value); err != nil {
		return "", handleNotFound(err)
	}
	return value, nil
}

func (s *RunBindingsStore) EnvironmentDefinitionExists(ctx context.Context, projectID, envDefID string) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("run bindings store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	envDefID = strings.TrimSpace(envDefID)
	if projectID == "" || envDefID == "" {
		return false, fmt.Errorf("project id and environment definition id are required")
	}
	row := s.db.QueryRowContext(
		ctx,
		selectEnvDefinitionExistsQuery,
		projectID,
		envDefID,
	)
	var v int
	if err := row.Scan(&v); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// nullIfEmpty is declared in runs.go within the same package.
