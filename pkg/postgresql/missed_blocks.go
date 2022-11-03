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
	CREATE_MISSED_BLOCKS_TABLE = `
		CREATE TABLE IF NOT EXISTS t_missed_blocks(
			f_slot INT,
			f_label TEXT,
			CONSTRAINT PK_MissedBlock PRIMARY KEY (f_slot,f_label));`

	InsertNewMissedBlock = `
		INSERT INTO t_missed_blocks (	
			f_slot, 
			f_label)
		VALUES ($1, $2);`
)

// in case the table did not exist
func (p *PostgresDBService) createMissedBlockMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CREATE_MISSED_BLOCKS_TABLE)
	if err != nil {
		return errors.Wrap(err, "error creating missed blocks metrics table")
	}
	return nil
}
