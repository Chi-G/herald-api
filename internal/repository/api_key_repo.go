package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"herald/internal/models"
)

type APIKeyRepository struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepository(db *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) FindByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, key_prefix, is_active, last_used_at, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true AND revoked_at IS NULL
	`
	var k models.APIKey
	err := r.db.QueryRow(ctx, query, keyHash).Scan(
		&k.ID,
		&k.TenantID,
		&k.Name,
		&k.KeyHash,
		&k.KeyPrefix,
		&k.IsActive,
		&k.LastUsedAt,
		&k.CreatedAt,
		&k.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find api key by hash: %w", err)
	}
	return &k, nil
}

func (r *APIKeyRepository) TouchLastUsed(ctx context.Context, keyID uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("touch api key last used: %w", err)
	}
	return nil
}
