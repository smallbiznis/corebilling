package domain

import "context"

// Repository defines persistence for the ledger.
type Repository interface {
	CreateAccount(ctx context.Context, account Account) error
	GetAccount(ctx context.Context, id string) (Account, error)
	ListAccounts(ctx context.Context, tenantID string) ([]Account, error)
	CreateJournalAndEntries(ctx context.Context, journal JournalEntry, entries []LedgerEntry) error
	Transfer(ctx context.Context, journal JournalEntry, entries []LedgerEntry) error
}
