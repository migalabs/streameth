package postgresql

/*

This file together with the model, has all the needed methods to interact with the epoch_metrics table of the database

*/

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

var (
	CreateAttTable = `
		CREATE TABLE IF NOT EXISTS t_att_metrics(
			f_label TEXT,
			f_slot INT,
			f_committee_index INT,
			f_signature TEXT,
			f_source_root TEXT,
			f_target_root TEXT,
			f_head_root TEXT,
			f_timestamp TIMESTAMP,
		CONSTRAINT PK_Attestation PRIMARY KEY (f_label,f_slot,f_committee_index,f_signature));`

	InsertNewAtt = `
		INSERT INTO t_att_metrics (	
			f_label, 
			f_slot,
			f_committee_index,
			f_timestamp,
			f_signature,
			f_source_root,
			f_target_root,
			f_head_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
