package domain

import "time"

// AccountType values.
const (
	AccountTypeUnspecified = 0
	AccountTypeCash        = 1
	AccountTypeInternal    = 2
	AccountTypeRevenue     = 3
	AccountTypeExpense     = 4
	AccountTypeLiability   = 5
	AccountTypeAsset       = 6
	AccountTypePointWallet = 7
)

// EntryType values.
const (
	EntryTypeUnspecified = 0
	EntryTypeDebit       = 1
	EntryTypeCredit      = 2
)

// Account represents a ledger account.
type Account struct {
	ID           string
	TenantID     string
	Name         string
	Type         int32
	Currency     string
	BalanceCents int64
	Metadata     map[string]interface{}
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// JournalEntry records metadata about a double-entry.
type JournalEntry struct {
	ID            string
	TenantID      string
	ReferenceID   string
	ReferenceType string
	Description   string
	Metadata      map[string]interface{}
	CreatedAt     time.Time
}

// LedgerEntry is a single debit/credit row.
type LedgerEntry struct {
	ID             string
	JournalEntryID string
	AccountID      string
	Type           int32
	AmountCents    int64
	CreatedAt      time.Time
}
