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
	CREATE_BLOCK_ARRIVAL_TABLE = `
		CREATE TABLE IF NOT EXISTS t_block_metrics(
			f_slot INT,
			f_label TEXT,
			f_timestamp TIME,
			CONSTRAINT PK_Block PRIMARY KEY (f_slot,f_label));`

	InsertNewBlock = `
		INSERT INTO t_block_metrics (	
			f_slot, 
			f_label, 
			f_timestamp)
		VALUES ($1, $2, $3);`
)

// in case the table did not exist
func (p *PostgresDBService) createBlockMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CREATE_BLOCK_ARRIVAL_TABLE)
	if err != nil {
		return errors.Wrap(err, "error creating block arrival metrics table")
	}
	return nil
}
