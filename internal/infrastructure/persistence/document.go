package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type documentPgRepository struct {
	db *sql.DB
}

func NewDocumentPgRepository(db *sql.DB) repos.DocumentRepository {
	return &documentPgRepository{db: db}
}

func (r *documentPgRepository) Create(ctx context.Context, doc entities.Document) error {
	query := `
		INSERT INTO documents (
			id, owner_type, owner_id, type, file_name, file_url, extension,
			is_active, uploaded_at, object_key
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, query,
		doc.ID,
		doc.OwnerType.String(),
		doc.OwnerID,
		doc.Type.String(),
		doc.FileName.String(),
		doc.FileURL.String(),
		doc.Extension.String(),
		doc.IsActive,
		doc.UploadedAt,
		doc.ObjectKey, // novo
	)
	return err
}

func (r *documentPgRepository) Update(ctx context.Context, doc entities.Document) error {
	query := `
		UPDATE documents SET
			owner_type=$1, owner_id=$2, type=$3,
			file_name=$4, file_url=$5, extension=$6,
			is_active=$7, uploaded_at=$8, object_key=$9
		WHERE id=$10`
	res, err := r.db.ExecContext(ctx, query,
		doc.OwnerType.String(),
		doc.OwnerID,
		doc.Type.String(),
		doc.FileName.String(),
		doc.FileURL.String(),
		doc.Extension.String(),
		doc.IsActive,
		doc.UploadedAt,
		doc.ObjectKey, // novo
		doc.ID,
	)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *documentPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *documentPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Document, error) {
	query := `
		SELECT id, owner_type, owner_id, type, file_name, file_url, extension, is_active, uploaded_at, object_key
		FROM documents WHERE id = $1`
	var d entities.Document

	// scan para temporários "primitivos" e depois converte para VOs
	var ownerType, typ, fileName, fileURL, ext, objectKey string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &ownerType, &d.OwnerID, &typ,
		&fileName, &fileURL, &ext,
		&d.IsActive, &d.UploadedAt, &objectKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Document{}, sql.ErrNoRows
		}
		return entities.Document{}, err
	}

	d.OwnerType = vos.DocumentOwnerType(ownerType)
	d.Type = vos.MustNewDocumentType(typ)
	d.FileName = vos.MustNewFilename(fileName)
	d.FileURL = vos.MustNewDocumentURL(fileURL)
	d.Extension = vos.MustNewFileExtension(ext)
	d.ObjectKey = objectKey

	return d, nil
}

func (r *documentPgRepository) FindAllByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]entities.Document, error) {
	query := `
		SELECT id, owner_type, owner_id, type, file_name, file_url, extension, is_active, uploaded_at, object_key
		FROM documents
		WHERE owner_id = $1
		ORDER BY type ASC, uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.Document
	for rows.Next() {
		var d entities.Document
		var ownerType, typ, fileName, fileURL, ext, objectKey string

		if err := rows.Scan(
			&d.ID, &ownerType, &d.OwnerID, &typ, &fileName, &fileURL, &ext, &d.IsActive, &d.UploadedAt, &objectKey,
		); err != nil {
			return nil, err
		}

		d.OwnerType = vos.DocumentOwnerType(ownerType)
		d.Type = vos.MustNewDocumentType(typ)
		d.FileName = vos.MustNewFilename(fileName)
		d.FileURL = vos.MustNewDocumentURL(fileURL)
		d.Extension = vos.MustNewFileExtension(ext)
		d.ObjectKey = objectKey

		list = append(list, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *documentPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

var _ repos.DocumentRepository = (*documentPgRepository)(nil)
