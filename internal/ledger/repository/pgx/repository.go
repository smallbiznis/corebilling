package pgx

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/ledger/domain"
)

// Repository interacts with PostgreSQL for ledger data.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a new ledger repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateAccount(ctx context.Context, account domain.Account) error {
	metadata, err := marshalJSON(account.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO ledger_accounts (
			id, tenant_id, name, type, currency,
			balance_cents, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, account.ID, account.TenantID, account.Name, account.Type, account.Currency, account.BalanceCents, metadata, account.CreatedAt, account.UpdatedAt)
	return err
}

func (r *Repository) GetAccount(ctx context.Context, id string) (domain.Account, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, type, currency,
		       balance_cents, metadata, created_at, updated_at
		FROM ledger_accounts
		WHERE id=$1
	`, id)
	var acc domain.Account
	var metadata []byte
	if err := row.Scan(
		&acc.ID,
		&acc.TenantID,
		&acc.Name,
		&acc.Type,
		&acc.Currency,
		&acc.BalanceCents,
		&metadata,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	); err != nil {
		return domain.Account{}, err
	}
	acc.Metadata = jsonToMap(metadata)
	return acc, nil
}

func (r *Repository) ListAccounts(ctx context.Context, tenantID string) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, type, currency,
		       balance_cents, metadata, created_at, updated_at
		FROM ledger_accounts
		WHERE tenant_id=$1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []domain.Account
	for rows.Next() {
		var acc domain.Account
		var metadata []byte
		if err := rows.Scan(
			&acc.ID,
			&acc.TenantID,
			&acc.Name,
			&acc.Type,
			&acc.Currency,
			&acc.BalanceCents,
			&metadata,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		acc.Metadata = jsonToMap(metadata)
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *Repository) CreateJournalAndEntries(ctx context.Context, journal domain.JournalEntry, entries []domain.LedgerEntry) error {
	return r.applyJournal(ctx, journal, entries)
}

func (r *Repository) Transfer(ctx context.Context, journal domain.JournalEntry, entries []domain.LedgerEntry) error {
	return r.applyJournal(ctx, journal, entries)
}

func (r *Repository) applyJournal(ctx context.Context, journal domain.JournalEntry, entries []domain.LedgerEntry) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	metadata, err := marshalJSON(journal.Metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO ledger_journals (
			id, tenant_id, reference_id, reference_type,
			description, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, journal.ID, journal.TenantID, journal.ReferenceID, journal.ReferenceType, journal.Description, metadata, journal.CreatedAt)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		_, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (
				id, journal_entry_id, account_id,
				entry_type, amount_cents, created_at
			) VALUES ($1,$2,$3,$4,$5,$6)
		`, entry.ID, entry.JournalEntryID, entry.AccountID, entry.Type, entry.AmountCents, entry.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func marshalJSON(value map[string]interface{}) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return json.Marshal(value)
}

func jsonToMap(value []byte) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	var dst map[string]interface{}
	if err := json.Unmarshal(value, &dst); err != nil {
		return nil
	}
	return dst
}

var _ domain.Repository = (*Repository)(nil)
