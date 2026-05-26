package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/mergerhq/merger/internal/domain"
	"github.com/mergerhq/merger/internal/events"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(25)

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Migrate(ctx context.Context) error {
	for _, statement := range postgresMigrations() {
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) SaveChangePacket(ctx context.Context, packet domain.ChangePacket) error {
	payload, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	query := `
insert into merger_change_packets (
  id, repo_full_name, pr_number, author_login, merge_lane, risk_score, decision_status, payload, created_at, updated_at
)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
on conflict (id) do update set
  merge_lane = excluded.merge_lane,
  risk_score = excluded.risk_score,
  decision_status = excluded.decision_status,
  payload = excluded.payload,
  updated_at = excluded.updated_at`

	_, err = r.db.ExecContext(
		ctx,
		query,
		packet.ID,
		packet.Repo.FullName,
		packet.PR.Number,
		packet.Author.Login,
		packet.MergeLane,
		packet.RiskSummary.Score,
		packet.Decision.Status,
		payload,
		packet.CreatedAt,
		packet.UpdatedAt,
	)

	return err
}

func (r *PostgresRepository) SaveEvent(ctx context.Context, event events.Envelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`insert into merger_event_log (id, event_type, source, correlation_id, causation_id, change_packet_id, payload, created_at)
		 values ($1,$2,$3,$4,$5,$6,$7,$8)
		 on conflict (id) do nothing`,
		event.ID,
		event.Type,
		event.Source,
		event.CorrelationID,
		event.CausationID,
		event.ChangePacketID,
		payload,
		time.Now().UTC(),
	)

	return err
}

func (r *PostgresRepository) Close() error {
	if r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("postgres repository not initialized")
	}
	return r.db.PingContext(ctx)
}
