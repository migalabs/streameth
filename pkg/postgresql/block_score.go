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
	CREATE_SCORE_TABLE = `
		CREATE TABLE IF NOT EXISTS t_score_metrics(
			f_slot INT,
			f_client_name TEXT,
			f_label TEXT,
			f_score FLOAT,
			f_duration FLOAT,
			f_correct_source INT,
			f_correct_target INT,
			f_correct_head INT,
			f_sync_bits INT,
			f_att_num INT,
			f_new_votes INT,
			f_attester_slashings INT,
			f_proposer_slashings INT,
			f_proposer_slashing_score FLOAT,
			f_attester_slashing_score FLOAT,
			f_sync_score FLOAT,
			CONSTRAINT PK_Score PRIMARY KEY (f_slot,f_label));`

	InsertNewScore = `
		INSERT INTO t_score_metrics (	
			f_slot, 
			f_client_name,
			f_label, 
			f_score,
			f_duration,
			f_correct_source,
			f_correct_target,
			f_correct_head,
			f_sync_bits,
			f_att_num,
			f_new_votes,
			f_attester_slashings,
			f_proposer_slashings,
			f_proposer_slashing_score,
			f_attester_slashing_score,
			f_sync_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);`
)

// in case the table did not exist
func (p *PostgresDBService) createScoreMetricsTable(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	_, err := pool.Exec(ctx, CREATE_SCORE_TABLE)
	if err != nil {
		return errors.Wrap(err, "error creating score metrics table")
	}
	return nil
}

type BlockMetricsModel struct {
	Slot                  int
	ClientName            string
	Label                 string
	Score                 float64
	Duration              float64
	CorrectSource         int
	CorrectTarget         int
	CorrectHead           int
	Sync1Bits             int
	AttNum                int
	NewVotes              int
	AttesterSlashings     int
	ProposerSlashings     int
	ProposerSlashingScore float64
	AttesterSlashingScore float64
	SyncScore             float64
}
