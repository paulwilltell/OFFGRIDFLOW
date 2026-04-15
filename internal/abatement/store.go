package abatement

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

type StoredEvidence struct {
	ID        string
	TenantID  string
	ActionID  string
	Framework Framework
	Path      string
	FileName  string
	MimeType  string
	SizeBytes int64
	Content   []byte
	CreatedAt time.Time
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SyncDefinitions(ctx context.Context, tenantID string, framework Framework, definitions []RiskDefinition) error {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return err
	}
	defer rollback(tx)

	query := `
INSERT INTO readiness_action_items (
  tenant_id, framework, compliance_check_id, title, severity, priority, description,
  acceptance_criteria, required_evidence_types, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
ON CONFLICT (tenant_id, framework, compliance_check_id)
DO UPDATE SET
  title = EXCLUDED.title,
  severity = EXCLUDED.severity,
  priority = EXCLUDED.priority,
  description = EXCLUDED.description,
  acceptance_criteria = EXCLUDED.acceptance_criteria,
  required_evidence_types = EXCLUDED.required_evidence_types,
  updated_at = NOW()`

	for _, def := range definitions {
		if _, err := tx.ExecContext(
			ctx,
			query,
			tenantID,
			string(framework),
			def.CheckID,
			def.Title,
			string(def.Severity),
			string(def.Priority),
			def.Description,
			pq.Array(def.AcceptanceCriteria),
			pq.Array(def.RequiredEvidenceTypes),
		); err != nil {
			return fmt.Errorf("sync readiness action item %s: %w", def.CheckID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sync readiness action items: %w", err)
	}
	return nil
}

func (s *Store) ListActionItems(ctx context.Context, tenantID string, framework Framework) ([]StoredActionItem, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	items, err := s.listActionItemsTx(ctx, tx, tenantID, framework)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit list action items: %w", err)
	}
	return items, nil
}

func (s *Store) GetActionItemByCheckID(ctx context.Context, tenantID string, framework Framework, checkID string) (*StoredActionItem, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	item, err := s.getActionItemByCheckIDTx(ctx, tx, tenantID, framework, checkID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit get action item: %w", err)
	}
	return item, nil
}

func (s *Store) SaveEvaluation(
	ctx context.Context,
	tenantID, userID string,
	framework Framework,
	itemID string,
	justification string,
	completed bool,
	evaluation Evaluation,
	evidence []EvidenceUpload,
) (*StoredActionItem, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	if _, err := s.getActionItemByIDTx(ctx, tx, tenantID, framework, itemID); err != nil {
		return nil, err
	}

	if len(evidence) > 0 {
		if err := s.insertEvidenceTx(ctx, tx, tenantID, userID, framework, itemID, evidence); err != nil {
			return nil, err
		}
	}

	evidenceURLs, err := s.listEvidenceURLsTx(ctx, tx, itemID, framework)
	if err != nil {
		return nil, err
	}

	query := `
UPDATE readiness_action_items
SET justification = $4,
    evidence_urls = $5,
    engine_status = $6,
    engine_feedback = $7,
    criteria_checked = $8,
    completed = $9,
    updated_by = $10,
    updated_at = NOW()
WHERE tenant_id = $1
  AND framework = $2
  AND id = $3`

	if _, err := tx.ExecContext(
		ctx,
		query,
		tenantID,
		string(framework),
		itemID,
		strings.TrimSpace(justification),
		pq.Array(evidenceURLs),
		nullEngineStatus(evaluation.Status),
		nullString(evaluation.Feedback),
		pq.Array(evaluation.CriteriaChecked),
		completed,
		nullUUID(userID),
	); err != nil {
		return nil, fmt.Errorf("update readiness action item evaluation: %w", err)
	}

	if err := s.logAuditEventTx(ctx, tx, tenantID, userID, "abatement.evaluate", itemID, map[string]any{
		"framework":         framework,
		"engine_status":     evaluation.Status,
		"criteria_checked":  evaluation.CriteriaChecked,
		"evidence_uploaded": len(evidence),
		"completed":         completed,
	}); err != nil {
		return nil, err
	}

	item, err := s.getActionItemByIDTx(ctx, tx, tenantID, framework, itemID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit evaluate action item: %w", err)
	}
	return item, nil
}

func (s *Store) SetSelfCertified(
	ctx context.Context,
	tenantID, userID string,
	framework Framework,
	actionItemID string,
	complianceCheckID string,
	selfCertified bool,
) (*StoredActionItem, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	itemID := strings.TrimSpace(actionItemID)
	if itemID == "" {
		item, err := s.getActionItemByCheckIDTx(ctx, tx, tenantID, framework, complianceCheckID)
		if err != nil {
			return nil, err
		}
		itemID = item.ID
	}

	query := `
UPDATE readiness_action_items
SET self_certified = $4,
    certified_at = CASE WHEN $4 THEN NOW() ELSE NULL END,
    certified_by = CASE WHEN $4 THEN $5::uuid ELSE NULL END,
    updated_by = $5::uuid,
    updated_at = NOW()
WHERE tenant_id = $1
  AND framework = $2
  AND id = $3`

	if _, err := tx.ExecContext(ctx, query, tenantID, string(framework), itemID, selfCertified, userID); err != nil {
		return nil, fmt.Errorf("update readiness action item self-certification: %w", err)
	}

	if err := s.logAuditEventTx(ctx, tx, tenantID, userID, "abatement.self_certify", itemID, map[string]any{
		"framework":       framework,
		"self_certified":  selfCertified,
		"compliance_check": complianceCheckID,
	}); err != nil {
		return nil, err
	}

	item, err := s.getActionItemByIDTx(ctx, tx, tenantID, framework, itemID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit self-certify action item: %w", err)
	}
	return item, nil
}

func (s *Store) GetEvidence(ctx context.Context, tenantID, evidenceID string, framework Framework) (*StoredEvidence, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	query := `
SELECT id, tenant_id, action_item_id, framework, storage_path, file_name, mime_type, file_size_bytes, file_bytes, created_at
FROM abatement_evidence
WHERE tenant_id = $1 AND id = $2 AND framework = $3`
	var record StoredEvidence
	if err := tx.QueryRowContext(ctx, query, tenantID, evidenceID, string(framework)).Scan(
		&record.ID,
		&record.TenantID,
		&record.ActionID,
		&record.Framework,
		&record.Path,
		&record.FileName,
		&record.MimeType,
		&record.SizeBytes,
		&record.Content,
		&record.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("abatement evidence not found")
		}
		return nil, fmt.Errorf("load abatement evidence: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit evidence lookup: %w", err)
	}

	return &record, nil
}

func (s *Store) beginTenantTx(ctx context.Context, tenantID string) (*sql.Tx, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("abatement store is not configured")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin abatement transaction: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("set tenant rls context: %w", err)
	}
	return tx, nil
}

func (s *Store) listActionItemsTx(ctx context.Context, tx *sql.Tx, tenantID string, framework Framework) ([]StoredActionItem, error) {
	query := `
SELECT id, tenant_id, framework, compliance_check_id, title, severity, priority, description,
       acceptance_criteria, required_evidence_types, COALESCE(justification, ''),
       evidence_urls, COALESCE(engine_status, ''), COALESCE(engine_feedback, ''),
       criteria_checked, completed, self_certified, certified_at, updated_at
FROM readiness_action_items
WHERE tenant_id = $1 AND framework = $2
ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, title`

	rows, err := tx.QueryContext(ctx, query, tenantID, string(framework))
	if err != nil {
		return nil, fmt.Errorf("list readiness action items: %w", err)
	}
	defer rows.Close()

	var items []StoredActionItem
	for rows.Next() {
		var item StoredActionItem
		var acceptance []string
		var requiredEvidence []string
		var evidenceURLs []string
		var criteriaChecked []string
		var rawStatus string
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.Framework,
			&item.ComplianceCheckID,
			&item.Title,
			&item.Severity,
			&item.Priority,
			&item.Description,
			pq.Array(&acceptance),
			pq.Array(&requiredEvidence),
			&item.Justification,
			pq.Array(&evidenceURLs),
			&rawStatus,
			&item.EngineFeedback,
			pq.Array(&criteriaChecked),
			&item.Completed,
			&item.SelfCertified,
			&item.CertifiedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan readiness action item: %w", err)
		}
		item.AcceptanceCriteria = acceptance
		item.RequiredEvidenceTypes = requiredEvidence
		item.EvidenceURLs = evidenceURLs
		item.CriteriaChecked = criteriaChecked
		item.EngineStatus = EngineStatus(rawStatus)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate readiness action items: %w", err)
	}
	return items, nil
}

func (s *Store) getActionItemByCheckIDTx(ctx context.Context, tx *sql.Tx, tenantID string, framework Framework, checkID string) (*StoredActionItem, error) {
	query := `
SELECT id, tenant_id, framework, compliance_check_id, title, severity, priority, description,
       acceptance_criteria, required_evidence_types, COALESCE(justification, ''),
       evidence_urls, COALESCE(engine_status, ''), COALESCE(engine_feedback, ''),
       criteria_checked, completed, self_certified, certified_at, updated_at
FROM readiness_action_items
WHERE tenant_id = $1 AND framework = $2 AND compliance_check_id = $3`
	return s.scanSingleActionItem(ctx, tx, query, tenantID, string(framework), checkID)
}

func (s *Store) getActionItemByIDTx(ctx context.Context, tx *sql.Tx, tenantID string, framework Framework, itemID string) (*StoredActionItem, error) {
	query := `
SELECT id, tenant_id, framework, compliance_check_id, title, severity, priority, description,
       acceptance_criteria, required_evidence_types, COALESCE(justification, ''),
       evidence_urls, COALESCE(engine_status, ''), COALESCE(engine_feedback, ''),
       criteria_checked, completed, self_certified, certified_at, updated_at
FROM readiness_action_items
WHERE tenant_id = $1 AND framework = $2 AND id = $3`
	return s.scanSingleActionItem(ctx, tx, query, tenantID, string(framework), itemID)
}

func (s *Store) scanSingleActionItem(ctx context.Context, tx *sql.Tx, query string, tenantID, framework, key string) (*StoredActionItem, error) {
	var item StoredActionItem
	var acceptance []string
	var requiredEvidence []string
	var evidenceURLs []string
	var criteriaChecked []string
	var rawStatus string

	if err := tx.QueryRowContext(ctx, query, tenantID, framework, key).Scan(
		&item.ID,
		&item.TenantID,
		&item.Framework,
		&item.ComplianceCheckID,
		&item.Title,
		&item.Severity,
		&item.Priority,
		&item.Description,
		pq.Array(&acceptance),
		pq.Array(&requiredEvidence),
		&item.Justification,
		pq.Array(&evidenceURLs),
		&rawStatus,
		&item.EngineFeedback,
		pq.Array(&criteriaChecked),
		&item.Completed,
		&item.SelfCertified,
		&item.CertifiedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("readiness action item not found")
		}
		return nil, fmt.Errorf("load readiness action item: %w", err)
	}

	item.AcceptanceCriteria = acceptance
	item.RequiredEvidenceTypes = requiredEvidence
	item.EvidenceURLs = evidenceURLs
	item.CriteriaChecked = criteriaChecked
	item.EngineStatus = EngineStatus(rawStatus)
	return &item, nil
}

func (s *Store) insertEvidenceTx(ctx context.Context, tx *sql.Tx, tenantID, userID string, framework Framework, itemID string, evidence []EvidenceUpload) error {
	query := `
INSERT INTO abatement_evidence (
  id, tenant_id, action_item_id, framework, storage_path, file_name, mime_type, file_size_bytes, file_bytes, created_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

	for _, file := range evidence {
		if len(file.Content) == 0 {
			continue
		}

		evidenceID := uuid.NewString()
		path := fmt.Sprintf("abatement/%s/%s/%s/%s-%s",
			tenantID,
			framework,
			itemID,
			evidenceID,
			safeFileName(file.FileName),
		)
		if _, err := tx.ExecContext(
			ctx,
			query,
			evidenceID,
			tenantID,
			itemID,
			string(framework),
			path,
			file.FileName,
			file.MimeType,
			len(file.Content),
			file.Content,
			nullUUID(userID),
		); err != nil {
			return fmt.Errorf("insert abatement evidence %s: %w", file.FileName, err)
		}
	}

	return nil
}

func (s *Store) listEvidenceURLsTx(ctx context.Context, tx *sql.Tx, itemID string, framework Framework) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id
FROM abatement_evidence
WHERE action_item_id = $1 AND framework = $2
ORDER BY created_at ASC`, itemID, string(framework))
	if err != nil {
		return nil, fmt.Errorf("list abatement evidence urls: %w", err)
	}
	defer rows.Close()

	urls := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan abatement evidence id: %w", err)
		}
		urls = append(urls, fmt.Sprintf("/api/abatement/%s/evidence/%s", framework, id))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate abatement evidence urls: %w", err)
	}
	return urls, nil
}

func (s *Store) ListEvidenceRecords(ctx context.Context, tenantID, actionItemID string, framework Framework) ([]EvidenceRecord, error) {
	tx, err := s.beginTenantTx(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	records, err := s.listEvidenceRecordsTx(ctx, tx, actionItemID, framework)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit evidence records lookup: %w", err)
	}
	return records, nil
}

func (s *Store) listEvidenceRecordsTx(ctx context.Context, tx *sql.Tx, actionItemID string, framework Framework) ([]EvidenceRecord, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, file_name, mime_type, file_size_bytes, created_at
FROM abatement_evidence
WHERE action_item_id = $1 AND framework = $2
ORDER BY created_at ASC`, actionItemID, string(framework))
	if err != nil {
		return nil, fmt.Errorf("query abatement evidence: %w", err)
	}
	defer rows.Close()

	records := make([]EvidenceRecord, 0)
	for rows.Next() {
		var record EvidenceRecord
		if err := rows.Scan(&record.ID, &record.FileName, &record.MimeType, &record.SizeBytes, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan abatement evidence: %w", err)
		}
		record.URL = fmt.Sprintf("/api/abatement/%s/evidence/%s", framework, record.ID)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate abatement evidence: %w", err)
	}
	return records, nil
}

func (s *Store) logAuditEventTx(ctx context.Context, tx *sql.Tx, tenantID, userID, action, entityID string, metadata map[string]any) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal abatement audit metadata: %w", err)
	}

	query := `
INSERT INTO audit_logs (tenant_id, user_id, action, entity_type, entity_id, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())`
	if _, err := tx.ExecContext(ctx, query, tenantID, nullUUID(userID), action, "readiness_action_item", entityID, payload); err != nil {
		return fmt.Errorf("insert abatement audit log: %w", err)
	}
	return nil
}

func rollback(tx *sql.Tx) {
	if tx != nil {
		_ = tx.Rollback()
	}
}

func safeFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		return "evidence.bin"
	}
	replacer := strings.NewReplacer(" ", "-", "\\", "-", "/", "-", ":", "-", ";", "-", "\"", "", "'", "", "..", ".")
	base = replacer.Replace(base)
	if base == "" {
		return "evidence.bin"
	}
	return base
}

func nullString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullEngineStatus(status EngineStatus) any {
	if status == "" {
		return nil
	}
	return string(status)
}

func nullUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
