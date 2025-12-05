package ledger

import (
	"context"
	"time"

	"github.com/smallbiznis/corebilling/internal/ledger/domain"
	repo "github.com/smallbiznis/corebilling/internal/ledger/repository/pgx"
	ledgerv1 "github.com/smallbiznis/go-genproto/smallbiznis/ledger/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Module wires ledger metrics.
var Module = fx.Options(
	fx.Provide(repo.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)

func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	ledgerv1.RegisterLedgerServiceServer(server, &grpcService{svc: svc})
}

type grpcService struct {
	ledgerv1.UnimplementedLedgerServiceServer
	svc *domain.Service
}

func (g *grpcService) CreateAccount(ctx context.Context, req *ledgerv1.CreateAccountRequest) (*ledgerv1.CreateAccountResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetName() == "" || req.GetCurrency() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant, name, and currency required")
	}
	account := domain.Account{
		TenantID: req.GetTenantId(),
		Name:     req.GetName(),
		Type:     int32(req.GetType()),
		Currency: req.GetCurrency(),
		Metadata: structToMap(req.GetMetadata()),
	}
	if err := g.svc.CreateAccount(ctx, account); err != nil {
		return nil, err
	}
	created, err := g.svc.GetAccount(ctx, account.ID)
	if err != nil {
		return nil, err
	}
	return &ledgerv1.CreateAccountResponse{Account: g.accountToProto(created)}, nil
}

func (g *grpcService) GetAccount(ctx context.Context, req *ledgerv1.GetAccountRequest) (*ledgerv1.GetAccountResponse, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	account, err := g.svc.GetAccount(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &ledgerv1.GetAccountResponse{Account: g.accountToProto(account)}, nil
}

func (g *grpcService) ListAccounts(ctx context.Context, req *ledgerv1.ListAccountsRequest) (*ledgerv1.ListAccountsResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id required")
	}
	accounts, err := g.svc.ListAccounts(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	resp := &ledgerv1.ListAccountsResponse{}
	for _, acc := range accounts {
		resp.Accounts = append(resp.Accounts, g.accountToProto(acc))
	}
	return resp, nil
}

func (g *grpcService) CreateJournalEntry(ctx context.Context, req *ledgerv1.CreateJournalEntryRequest) (*ledgerv1.CreateJournalEntryResponse, error) {
	if err := validateJournalRequest(req); err != nil {
		return nil, err
	}
	journal := domain.JournalEntry{
		TenantID:      req.GetTenantId(),
		ReferenceID:   req.GetReferenceId(),
		ReferenceType: req.GetReferenceType(),
		Description:   req.GetDescription(),
		Metadata:      structToMap(req.GetMetadata()),
	}
	entries := make([]domain.LedgerEntry, 0, len(req.GetLines()))
	for _, line := range req.GetLines() {
		entries = append(entries, domain.LedgerEntry{
			AccountID:   line.GetAccountId(),
			Type:        int32(line.GetType()),
			AmountCents: line.GetAmountCents(),
		})
	}
	if err := g.svc.CreateJournalEntry(ctx, journal, entries); err != nil {
		return nil, err
	}
	return &ledgerv1.CreateJournalEntryResponse{
		Journal: g.journalToProto(journal),
	}, nil
}

func (g *grpcService) Transfer(ctx context.Context, req *ledgerv1.TransferRequest) (*ledgerv1.TransferResponse, error) {
	if err := validateTransferRequest(req); err != nil {
		return nil, err
	}
	journal := domain.JournalEntry{
		TenantID:      req.GetTenantId(),
		ReferenceID:   req.GetReferenceId(),
		ReferenceType: req.GetReferenceType(),
		Description:   req.GetDescription(),
		Metadata:      structToMap(req.GetMetadata()),
	}
	now := time.Now().UTC()
	entries := []domain.LedgerEntry{
		{
			AccountID:   req.GetFromAccountId(),
			Type:        domain.EntryTypeCredit,
			AmountCents: req.GetAmountCents(),
			CreatedAt:   now,
		},
		{
			AccountID:   req.GetToAccountId(),
			Type:        domain.EntryTypeDebit,
			AmountCents: req.GetAmountCents(),
			CreatedAt:   now,
		},
	}
	if err := g.svc.Transfer(ctx, journal, entries); err != nil {
		return nil, err
	}
	respEntries := make([]*ledgerv1.LedgerEntry, len(entries))
	for i, entry := range entries {
		respEntries[i] = g.entryToProto(entry)
	}
	return &ledgerv1.TransferResponse{
		Result: &ledgerv1.TransferResult{
			Journal: g.journalToProto(journal),
			Entries: respEntries,
		},
	}, nil
}

func validateJournalRequest(req *ledgerv1.CreateJournalEntryRequest) error {
	if req == nil || req.GetTenantId() == "" || len(req.GetLines()) == 0 {
		return status.Error(codes.InvalidArgument, "tenant_id and at least one line required")
	}
	return nil
}

func validateTransferRequest(req *ledgerv1.TransferRequest) error {
	if req == nil || req.GetTenantId() == "" || req.GetFromAccountId() == "" || req.GetToAccountId() == "" || req.GetAmountCents() <= 0 {
		return status.Error(codes.InvalidArgument, "tenant, accounts and positive amount required")
	}
	return nil
}

func (g *grpcService) accountToProto(acc domain.Account) *ledgerv1.Account {
	return &ledgerv1.Account{
		Id:           acc.ID,
		TenantId:     acc.TenantID,
		Name:         acc.Name,
		Type:         ledgerv1.AccountType(acc.Type),
		Currency:     acc.Currency,
		BalanceCents: acc.BalanceCents,
		Metadata:     mapToStruct(acc.Metadata),
		CreatedAt:    timestamppb.New(acc.CreatedAt),
		UpdatedAt:    timestamppb.New(acc.UpdatedAt),
	}
}

func (g *grpcService) journalToProto(j domain.JournalEntry) *ledgerv1.JournalEntry {
	return &ledgerv1.JournalEntry{
		Id:            j.ID,
		TenantId:      j.TenantID,
		ReferenceId:   j.ReferenceID,
		ReferenceType: j.ReferenceType,
		Description:   j.Description,
		Metadata:      mapToStruct(j.Metadata),
		CreatedAt:     timestamppb.New(j.CreatedAt),
	}
}

func (g *grpcService) entryToProto(e domain.LedgerEntry) *ledgerv1.LedgerEntry {
	return &ledgerv1.LedgerEntry{
		Id:             e.ID,
		JournalEntryId: e.JournalEntryID,
		AccountId:      e.AccountID,
		Type:           ledgerv1.EntryType(e.Type),
		AmountCents:    e.AmountCents,
		CreatedAt:      timestamppb.New(e.CreatedAt),
	}
}

func structToMap(value *structpb.Struct) map[string]interface{} {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func mapToStruct(value map[string]interface{}) *structpb.Struct {
	if len(value) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(value)
	if err != nil {
		return nil
	}
	return s
}
