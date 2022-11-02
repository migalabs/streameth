package postgresql

/*

This file together with the model, has all the needed methods to interact with the epoch_metrics table of the database

*/

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

var (
	CreateAttTable = `
		CREATE TABLE IF NOT EXISTS t_att_metrics(
			f_label VARCHAR(100),
			f_slot INT,
			f_committee_index INT,
			f_val_idx INT,
			f_timestamp TIME,
			CONSTRAINT PK_Attestation PRIMARY KEY (f_label,f_slot,f_committee_index,f_val_idx));`

	InsertNewAtt = `
		INSERT INTO t_att_metrics (	
			f_label, 
			f_slot,
			f_committee_index,
			f_timestamp,
			f_val_idx 
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING;`
)

// in case the table did not exist
func (p *PostgresDBService) createAttMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CreateAttTable)
	if err != nil {
		return errors.Wrap(err, "error creating attestation metrics table")
	}
	return nil
}

func (p *PostgresDBService) InsertNewAtt(slot int, label string, timestamp time.Time) error {

	_, err := p.psqlPool.Exec(p.ctx, InsertNewAtt,
		slot,
		label,
		timestamp)

	if err != nil {
		return errors.Wrap(err, "error inserting row in score metrics table")
	}
	return nil
}
