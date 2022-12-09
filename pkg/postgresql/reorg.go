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
	CreateReorgTable = `
		CREATE TABLE IF NOT EXISTS t_reorg_metrics(
			f_label TEXT,
			f_slot INT,
			f_old_head TEXT,
			f_new_head TEXT,
			f_depth INT,
			f_timestamp TIMESTAMP,
		CONSTRAINT PK_Reorg PRIMARY KEY (f_label,f_slot));`

	InsertNewReorg = `
		INSERT INTO t_reorg_metrics (	
			f_label, 
			f_slot,
			f_old_head,
			f_new_head,
			f_depth,
			f_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING;`
)

// in case the table did not exist
func (p *PostgresDBService) createReorgMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CreateReorgTable)
	if err != nil {
		return errors.Wrap(err, "error creating reorgs metrics table")
	}
	return nil
}
