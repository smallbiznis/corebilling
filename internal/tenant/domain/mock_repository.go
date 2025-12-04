package domain

import (
	"context"
)

// MockRepository is a hand-written mock compatible with tests which expect
// the `go.uber.org/mock` style controller and `EXPECT()` builder.
type MockRepository struct {
	ctrl interface{}

	// Last captured arguments
	LastCreateOwner Tenant
	LastUpdated     Tenant
	LastListFilter  ListFilter

	// configured returns for expected calls
	expectCreate  *mockCall
	expectGetByID *mockCall
	expectList    *mockCall
	expectUpdate  *mockCall
}

// mockCall holds return values configured via .Return(...)
type mockCall struct {
	returns []interface{}
}

// Return sets return values for the expectation.
func (m *mockCall) Return(vals ...interface{}) {
	m.returns = vals
}

// MockRepositoryMockRecorder provides the EXPECT() API used in tests.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository constructs a mock tied to the provided controller.
func NewMockRepository(ctrl interface{}) *MockRepository {
	return &MockRepository{ctrl: ctrl}
}

// EXPECT returns the recorder for configuring expected calls.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return &MockRepositoryMockRecorder{mock: m}
}

// Create expectation builder
func (r *MockRepositoryMockRecorder) Create() *mockCall {
	r.mock.expectCreate = &mockCall{}
	return r.mock.expectCreate
}

// GetByID expectation builder
func (r *MockRepositoryMockRecorder) GetByID() *mockCall {
	r.mock.expectGetByID = &mockCall{}
	return r.mock.expectGetByID
}

// List expectation builder
func (r *MockRepositoryMockRecorder) List() *mockCall {
	r.mock.expectList = &mockCall{}
	return r.mock.expectList
}

// Update expectation builder
func (r *MockRepositoryMockRecorder) Update() *mockCall {
	r.mock.expectUpdate = &mockCall{}
	return r.mock.expectUpdate
}

// Create records the call and returns configured result (if any).
func (m *MockRepository) Create(ctx context.Context, tenant Tenant) error {
	m.LastCreateOwner = tenant
	if m.expectCreate != nil && len(m.expectCreate.returns) > 0 {
		if err, ok := m.expectCreate.returns[0].(error); ok {
			return err
		}
	}
	return nil
}

// GetByID records the call and returns configured result (if any).
func (m *MockRepository) GetByID(ctx context.Context, id string) (Tenant, error) {
	if m.expectGetByID != nil && len(m.expectGetByID.returns) > 0 {
		t, _ := m.expectGetByID.returns[0].(Tenant)
		var err error
		if len(m.expectGetByID.returns) > 1 {
			err, _ = m.expectGetByID.returns[1].(error)
		}
		return t, err
	}
	return Tenant{}, nil
}

// List records the call and returns configured result (if any).
func (m *MockRepository) List(ctx context.Context, filter ListFilter) ([]Tenant, error) {
	m.LastListFilter = filter
	if m.expectList != nil && len(m.expectList.returns) > 0 {
		if tenants, ok := m.expectList.returns[0].([]Tenant); ok {
			var err error
			if len(m.expectList.returns) > 1 {
				err, _ = m.expectList.returns[1].(error)
			}
			return tenants, err
		}
	}
	return nil, nil
}

// Update records the call and returns configured result (if any).
func (m *MockRepository) Update(ctx context.Context, tenant Tenant) error {
	m.LastUpdated = tenant
	if m.expectUpdate != nil && len(m.expectUpdate.returns) > 0 {
		if err, ok := m.expectUpdate.returns[0].(error); ok {
			return err
		}
	}
	return nil
}
