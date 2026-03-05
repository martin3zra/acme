package app

import (
	"context"
	"database/sql"
	"errors"
	"time"

	guuid "github.com/google/uuid"
)

// findTemplateByID retrieves a template by ID with company isolation
func (s *Server) findTemplateByID(ctx context.Context, templateID int) (*Template, error) {
	var t Template
	var createdAt, updatedAt time.Time
	var deletedAt sql.NullTime

	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, uuid, company_id, name, description, status, current_version_id, created_at, updated_at, deleted_at
		 FROM templates
		 WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL`,
		templateID,
		CurrentCompany(ctx).ID,
	).Scan(
		&t.ID, &t.UUID, &t.CompanyID, &t.Name, &t.Description, &t.Status, &t.CurrentVersionID,
		&createdAt, &updatedAt, &deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	t.CreatedAt = &createdAt
	t.UpdatedAt = &updatedAt
	if deletedAt.Valid {
		t.DeletedAt = &deletedAt.Time
	}

	return &t, nil
}

// findTemplateVersionByID retrieves a template version
func (s *Server) findTemplateVersionByID(ctx context.Context, versionID int) (*TemplateVersion, error) {
	var tv TemplateVersion
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(
		ctx,
		`SELECT tv.id, tv.uuid, tv.template_id, tv.version_number, tv.layout_json, tv.status, tv.notes, tv.created_at, tv.updated_at
		 FROM template_versions tv
		 INNER JOIN templates t ON tv.template_id = t.id
		 WHERE tv.id = $1 AND t.company_id = $2`,
		versionID,
		CurrentCompany(ctx).ID,
	).Scan(
		&tv.ID, &tv.UUID, &tv.TemplateID, &tv.VersionNumber, &tv.LayoutJSON, &tv.Status, &tv.Notes,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("template version not found")
		}
		return nil, err
	}

	tv.CreatedAt = &createdAt
	tv.UpdatedAt = &updatedAt

	return &tv, nil
}

// listTemplates retrieves all templates for a company
func (s *Server) listTemplates(ctx context.Context) ([]*Template, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, uuid, company_id, name, description, status, current_version_id, created_at, updated_at, deleted_at
		 FROM templates
		 WHERE company_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC`,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*Template
	for rows.Next() {
		var t Template
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.UUID, &t.CompanyID, &t.Name, &t.Description, &t.Status, &t.CurrentVersionID,
			&createdAt, &updatedAt, &deletedAt,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt = &createdAt
		t.UpdatedAt = &updatedAt
		if deletedAt.Valid {
			t.DeletedAt = &deletedAt.Time
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// getPublishedVersion retrieves the published version of a template
func (s *Server) getPublishedVersion(ctx context.Context, templateID int) (*TemplateVersion, error) {
	var tv TemplateVersion
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(
		ctx,
		`SELECT tv.id, tv.uuid, tv.template_id, tv.version_number, tv.layout_json, tv.status, tv.notes, tv.created_at, tv.updated_at
		 FROM template_versions tv
		 INNER JOIN templates t ON tv.template_id = t.id
		 WHERE tv.template_id = $1 AND tv.status = 'published' AND t.company_id = $2
		 ORDER BY tv.version_number DESC
		 LIMIT 1`,
		templateID,
		CurrentCompany(ctx).ID,
	).Scan(
		&tv.ID, &tv.UUID, &tv.TemplateID, &tv.VersionNumber, &tv.LayoutJSON, &tv.Status, &tv.Notes,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no published version found for this template")
		}
		return nil, err
	}

	tv.CreatedAt = &createdAt
	tv.UpdatedAt = &updatedAt

	return &tv, nil
}

// listTemplateVersions retrieves all versions of a template
func (s *Server) listTemplateVersions(ctx context.Context, templateID int) ([]*TemplateVersion, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT tv.id, tv.uuid, tv.template_id, tv.version_number, tv.layout_json, tv.status, tv.notes, tv.created_at, tv.updated_at
		 FROM template_versions tv
		 INNER JOIN templates t ON tv.template_id = t.id
		 WHERE tv.template_id = $1 AND t.company_id = $2
		 ORDER BY tv.version_number DESC`,
		templateID,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*TemplateVersion
	for rows.Next() {
		var tv TemplateVersion
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&tv.ID, &tv.UUID, &tv.TemplateID, &tv.VersionNumber, &tv.LayoutJSON, &tv.Status, &tv.Notes,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		tv.CreatedAt = &createdAt
		tv.UpdatedAt = &updatedAt

		versions = append(versions, &tv)
	}

	return versions, rows.Err()
}

// storeTemplate creates a new template with initial draft version
func (s *Server) storeTemplate(ctx context.Context, form *StoreTemplateForm) (*Template, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	uuid := guuid.New().String()

	// Insert template
	var templateID int
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO templates (uuid, company_id, name, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 RETURNING id`,
		uuid,
		CurrentCompany(ctx).ID,
		form.Name,
		form.Description,
	).Scan(&templateID)
	if err != nil {
		return nil, err
	}

	// Create initial draft version
	versionUUID := guuid.New().String()
	var versionID int
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO template_versions (uuid, template_id, version_number, layout_json, status, created_at, updated_at)
		 VALUES ($1, $2, 1, $3, 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 RETURNING id`,
		versionUUID,
		templateID,
		form.LayoutJSON,
	).Scan(&versionID)
	if err != nil {
		return nil, err
	}

	// Update template current_version_id
	_, err = tx.ExecContext(
		ctx,
		`UPDATE templates SET current_version_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		versionID,
		templateID,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &Template{
		ID:               templateID,
		UUID:             uuid,
		CompanyID:        CurrentCompany(ctx).ID,
		Name:             form.Name,
		Description:      form.Description,
		Status:           "draft",
		CurrentVersionID: &versionID,
	}, nil
}

// updateTemplate updates template draft and creates new version
func (s *Server) updateTemplate(ctx context.Context, templateID int, form *StoreTemplateForm) (*Template, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Verify ownership and get current version number
	var currentVersionNumber int
	err = tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(MAX(tv.version_number), 0)
		 FROM template_versions tv
		 INNER JOIN templates t ON tv.template_id = t.id
		 WHERE tv.template_id = $1 AND t.company_id = $2`,
		templateID,
		CurrentCompany(ctx).ID,
	).Scan(&currentVersionNumber)
	if err != nil {
		return nil, err
	}

	// Update template metadata
	_, err = tx.ExecContext(
		ctx,
		`UPDATE templates SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $3 AND company_id = $4`,
		form.Name,
		form.Description,
		templateID,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return nil, err
	}

	// Create new draft version
	newVersionNumber := currentVersionNumber + 1
	versionUUID := guuid.New().String()
	var versionID int
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO template_versions (uuid, template_id, version_number, layout_json, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 RETURNING id`,
		versionUUID,
		templateID,
		newVersionNumber,
		form.LayoutJSON,
	).Scan(&versionID)
	if err != nil {
		return nil, err
	}

	// Update template current_version_id
	_, err = tx.ExecContext(
		ctx,
		`UPDATE templates SET current_version_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		versionID,
		templateID,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	template, err := s.findTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return template, nil
}

// publishTemplate publishes a template version
func (s *Server) publishTemplate(ctx context.Context, templateID int) (*TemplateVersion, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the current draft version
	var versionID int
	var versionNumber int
	err = tx.QueryRowContext(
		ctx,
		`SELECT tv.id, tv.version_number
		 FROM template_versions tv
		 INNER JOIN templates t ON tv.template_id = t.id
		 WHERE tv.template_id = $1 AND tv.status = 'draft' AND t.company_id = $2
		 ORDER BY tv.version_number DESC
		 LIMIT 1`,
		templateID,
		CurrentCompany(ctx).ID,
	).Scan(&versionID, &versionNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no draft version found to publish")
		}
		return nil, err
	}

	// Unpublish previous version
	_, err = tx.ExecContext(
		ctx,
		`UPDATE template_versions SET status = 'draft' WHERE template_id = $1 AND status = 'published'`,
		templateID,
	)
	if err != nil {
		return nil, err
	}

	// Publish new version
	_, err = tx.ExecContext(
		ctx,
		`UPDATE template_versions SET status = 'published', updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		versionID,
	)
	if err != nil {
		return nil, err
	}

	// Update template status
	_, err = tx.ExecContext(
		ctx,
		`UPDATE templates SET status = 'published', updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		templateID,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return s.findTemplateVersionByID(ctx, versionID)
}

// deleteTemplate soft deletes a template
func (s *Server) deleteTemplate(ctx context.Context, templateID int) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE templates SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND company_id = $2`,
		templateID,
		CurrentCompany(ctx).ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("template not found")
	}

	return nil
}

// duplicateTemplate creates a copy of a template with a new name
func (s *Server) duplicateTemplate(ctx context.Context, templateID int, newName string) (*Template, error) {
	// Get source template with current version
	source, err := s.findTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if source.CurrentVersionID == nil {
		return nil, errors.New("source template has no version")
	}

	version, err := s.findTemplateVersionByID(ctx, *source.CurrentVersionID)
	if err != nil {
		return nil, err
	}

	// Create new template with same layout
	form := &StoreTemplateForm{
		Name:        newName,
		Description: source.Description,
		LayoutJSON:  version.LayoutJSON,
	}

	return s.storeTemplate(ctx, form)
}
