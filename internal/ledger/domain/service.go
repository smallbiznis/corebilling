package domain

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Service contains ledger operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs ledger service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.Named("ledger.service"),
	}
}

// CreateAccount stores an account.
func (s *Service) CreateAccount(ctx context.Context, account Account) error {
	account.ID = account.ID
	if account.ID == "" {
		account.ID = ulid.Make().String()
	}
	now := time.Now().UTC()
	account.CreatedAt = now
	account.UpdatedAt = now
	return s.repo.CreateAccount(ctx, account)
}

// GetAccount fetches an account.
func (s *Service) GetAccount(ctx context.Context, id string) (Account, error) {
	return s.repo.GetAccount(ctx, id)
}

// ListAccounts lists tenant accounts.
func (s *Service) ListAccounts(ctx context.Context, tenantID string) ([]Account, error) {
	return s.repo.ListAccounts(ctx, tenantID)
}

// CreateJournalEntry posts a journal.
func (s *Service) CreateJournalEntry(ctx context.Context, journal JournalEntry, entries []LedgerEntry) error {
	if err := validateEntries(entries); err != nil {
		return err
	}
	if journal.ID == "" {
		journal.ID = ulid.Make().String()
	}
	journal.CreatedAt = time.Now().UTC()
	for i := range entries {
		entries[i].ID = ulid.Make().String()
		entries[i].JournalEntryID = journal.ID
		entries[i].CreatedAt = time.Now().UTC()
	}
	return s.repo.CreateJournalAndEntries(ctx, journal, entries)
}

// Transfer money between accounts.
func (s *Service) Transfer(ctx context.Context, journal JournalEntry, entries []LedgerEntry) error {
	return s.CreateJournalEntry(ctx, journal, entries)
}

func validateEntries(entries []LedgerEntry) error {
	var sum int64
	for _, entry := range entries {
		switch entry.Type {
		case EntryTypeDebit:
			sum += entry.AmountCents
		case EntryTypeCredit:
			sum -= entry.AmountCents
		default:
			return errors.New("invalid entry type")
		}
	}
	if sum != 0 {
		return errors.New("entries must balance")
	}
	return nil
}

func structToMap(value *structpb.Struct) map[string]interface{} {
	if value == nil {
		return nil
	}
	return value.AsMap()
}
