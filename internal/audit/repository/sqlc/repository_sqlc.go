package sqlc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	auditv1 "github.com/smallbiznis/go-genproto/smallbiznis/audit/v1"

	"github.com/smallbiznis/corebilling/internal/audit/domain"
)

// Repository stores audit events.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an audit repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, event domain.AuditEvent) error {
	oldValues, err := marshalJSON(event.OldValues)
	if err != nil {
		return err
	}
	newValues, err := marshalJSON(event.NewValues)
	if err != nil {
		return err
	}
	metadata, err := marshalJSON(event.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO audit_events (
			id, tenant_id, actor_type, actor_id, action, action_type, resource_type,
			resource_id, old_values, new_values, ip_address, user_agent, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, event.ID, event.TenantID, int16(event.ActorType), event.ActorID, event.Action, int16(event.ActionType),
		event.ResourceType, event.ResourceID, oldValues, newValues, event.IpAddress, event.UserAgent, metadata, event.CreatedAt)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.AuditEvent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, actor_type, actor_id, action, action_type, resource_type,
		       resource_id, old_values, new_values, ip_address, user_agent, metadata, created_at
		FROM audit_events WHERE id=$1
	`, id)

	var event domain.AuditEvent
	var actorType, actionType int16
	var oldValues, newValues, metadata []byte
	if err := row.Scan(
		&event.ID,
		&event.TenantID,
		&actorType,
		&event.ActorID,
		&event.Action,
		&actionType,
		&event.ResourceType,
		&event.ResourceID,
		&oldValues,
		&newValues,
		&event.IpAddress,
		&event.UserAgent,
		&metadata,
		&event.CreatedAt,
	); err != nil {
		return domain.AuditEvent{}, err
	}
	event.ActorType = auditv1.ActorType(actorType)
	event.ActionType = auditv1.ActionType(actionType)
	event.OldValues = jsonToMap(oldValues)
	event.NewValues = jsonToMap(newValues)
	event.Metadata = jsonToMap(metadata)
	return event, nil
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]domain.AuditEvent, error) {
	args := make([]interface{}, 0, 8)
	conds := make([]string, 0, 6)
	argIdx := 1

	if filter.TenantID != "" {
		args = append(args, filter.TenantID)
		conds = append(conds, fmt.Sprintf("tenant_id=$%d", argIdx))
		argIdx++
	}
	if filter.ResourceType != "" {
		args = append(args, filter.ResourceType)
		conds = append(conds, fmt.Sprintf("resource_type=$%d", argIdx))
		argIdx++
	}
	if filter.ResourceID != "" {
		args = append(args, filter.ResourceID)
		conds = append(conds, fmt.Sprintf("resource_id=$%d", argIdx))
		argIdx++
	}
	if filter.ActorID != "" {
		args = append(args, filter.ActorID)
		conds = append(conds, fmt.Sprintf("actor_id=$%d", argIdx))
		argIdx++
	}
	if filter.ActionType != auditv1.ActionType_ACTION_TYPE_UNSPECIFIED {
		args = append(args, int16(filter.ActionType))
		conds = append(conds, fmt.Sprintf("action_type=$%d", argIdx))
		argIdx++
	}
	if len(conds) == 0 {
		conds = append(conds, "TRUE")
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	limitIdx := argIdx
	args = append(args, filter.Limit)
	argIdx++
	offsetIdx := argIdx
	args = append(args, filter.Offset)
	argIdx++

	var builder strings.Builder
	builder.WriteString(`
		SELECT id, tenant_id, actor_type, actor_id, action, action_type, resource_type,
		       resource_id, old_values, new_values, ip_address, user_agent, metadata, created_at
		FROM audit_events WHERE `)
	builder.WriteString(strings.Join(conds, " AND "))
	builder.WriteString(` ORDER BY created_at DESC`)
	builder.WriteString(fmt.Sprintf(" LIMIT $%d", limitIdx))
	builder.WriteString(fmt.Sprintf(" OFFSET $%d", offsetIdx))

	rows, err := r.pool.Query(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var actorType, actionType int16
		var oldValues, newValues, metadata []byte
		if err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&actorType,
			&event.ActorID,
			&event.Action,
			&actionType,
			&event.ResourceType,
			&event.ResourceID,
			&oldValues,
			&newValues,
			&event.IpAddress,
			&event.UserAgent,
			&metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		event.ActorType = auditv1.ActorType(actorType)
		event.ActionType = auditv1.ActionType(actionType)
		event.OldValues = jsonToMap(oldValues)
		event.NewValues = jsonToMap(newValues)
		event.Metadata = jsonToMap(metadata)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
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
	var data map[string]interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return nil
	}
	return data
}

var _ domain.Repository = (*Repository)(nil)
