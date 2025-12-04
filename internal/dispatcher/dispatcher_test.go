package dispatcher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/smallbiznis/corebilling/internal/dispatcher"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	busmock "github.com/smallbiznis/corebilling/internal/mocks/bus"
	outboxmock "github.com/smallbiznis/corebilling/internal/mocks/outbox"
)

func TestDispatcher_Process_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ctrl *gomock.Controller) (*outboxmock.MockOutboxRepository, *busmock.MockEventBus)
		wantErr bool
	}{
		{
			name: "success single event",
			setup: func(ctrl *gomock.Controller) (*outboxmock.MockOutboxRepository, *busmock.MockEventBus) {
				or := outboxmock.NewMockOutboxRepository(ctrl)
				bb := busmock.NewMockEventBus(ctrl)

				ev := outbox.OutboxEvent{ID: "evt_1", Subject: "s1", Payload: []byte("p")}
				or.EXPECT().FetchPendingEvents(gomock.Any(), int32(100), gomock.Any()).Return([]outbox.OutboxEvent{ev}, nil).Times(1)
				bb.EXPECT().Publish(gomock.Any(), "s1", []byte("p")).Return(nil).Times(1)
				or.EXPECT().MarkDispatched(gomock.Any(), "evt_1").Return(nil).Times(1)
				return or, bb
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setup: func(ctrl *gomock.Controller) (*outboxmock.MockOutboxRepository, *busmock.MockEventBus) {
				or := outboxmock.NewMockOutboxRepository(ctrl)
				bb := busmock.NewMockEventBus(ctrl)
				or.EXPECT().FetchPendingEvents(gomock.Any(), int32(100), gomock.Any()).Return(nil, errors.New("db error")).Times(1)
				bb.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				return or, bb
			},
			wantErr: true,
		},
		{
			name: "bus error",
			setup: func(ctrl *gomock.Controller) (*outboxmock.MockOutboxRepository, *busmock.MockEventBus) {
				or := outboxmock.NewMockOutboxRepository(ctrl)
				bb := busmock.NewMockEventBus(ctrl)
				ev := outbox.OutboxEvent{ID: "e2", Subject: "s2", Payload: []byte("x")}
				or.EXPECT().FetchPendingEvents(gomock.Any(), int32(100), gomock.Any()).Return([]outbox.OutboxEvent{ev}, nil).Times(1)
				bb.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("publish error")).Times(1)
				or.EXPECT().MarkDispatched(gomock.Any(), gomock.Any()).Times(0)
				return or, bb
			},
			wantErr: true,
		},
		{
			name: "multiple events times verification",
			setup: func(ctrl *gomock.Controller) (*outboxmock.MockOutboxRepository, *busmock.MockEventBus) {
				or := outboxmock.NewMockOutboxRepository(ctrl)
				bb := busmock.NewMockEventBus(ctrl)
				ev1 := outbox.OutboxEvent{ID: "a", Subject: "s", Payload: []byte("1")}
				ev2 := outbox.OutboxEvent{ID: "b", Subject: "s", Payload: []byte("2")}
				or.EXPECT().FetchPendingEvents(gomock.Any(), int32(100), gomock.Any()).Return([]outbox.OutboxEvent{ev1, ev2}, nil).Times(1)
				bb.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
				or.EXPECT().MarkDispatched(gomock.Any(), "a").Return(nil).Times(1)
				or.EXPECT().MarkDispatched(gomock.Any(), "b").Return(nil).Times(1)
				return or, bb
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo, bus := tc.setup(ctrl)
			d := dispatcher.NewDispatcher(repo, bus)
			err := d.Process(context.Background(), 100)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Process() error = %v, wantErr= %v", err, tc.wantErr)
			}
		})
	}
}

func TestDispatcher_InsertEvent_ArgumentMatcher(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	or := outboxmock.NewMockOutboxRepository(ctrl)
	or.EXPECT().InsertOutboxEvent(gomock.Any(), gomock.AssignableToTypeOf(&outbox.OutboxEvent{})).Return(nil).Times(1)

	// call with a concrete outbox event
	err := or.InsertOutboxEvent(context.Background(), &outbox.OutboxEvent{ID: "x", Subject: "s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
