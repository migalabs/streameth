package postgresql

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pgx_v4 "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Static postgres queries, for each modification in the tables, the table needs to be reseted
var (
	// logrus associated with the postgres db
	PsqlType = "postgres-db"
	log      = logrus.WithField(
		"module", PsqlType,
	)
	MAX_BATCH_QUEUE   = 100
	WRITE_CHAN_LENGTH = 400000
)

type PostgresDBService struct {
	// Control Variables
	ctx              context.Context
	cancel           context.CancelFunc
	connectionUrl    string // the url might not be necessary (better to remove it?¿)
	psqlPool         *pgxpool.Pool
	WriteChan        chan WriteTask
	doneTasks        chan interface{}
	endProcess       int32
	FinishSignalChan chan struct{}
	workerNum        int
	maxBatchQueue    int
	WgDBWriter       sync.WaitGroup
}

// Connect to the PostgreSQL Database and get the multithread-proof connection
// from the given url-composed credentials
func ConnectToDB(ctx context.Context, url string, workerNum int, batchLen int) (*PostgresDBService, error) {
	mainCtx, cancel := context.WithCancel(ctx)
	// spliting the url to don't share any confidential information on logs

	if strings.Contains(url, "@") {
		log.Debugf("Connecting to PostgresDB at %s", strings.Split(url, "@")[1])
	}
	psqlPool, err := pgxpool.Connect(mainCtx, url)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "@") {
		log.Infof("PostgresDB %s succesfully connected", strings.Split(url, "@")[1])
	}
	// filter the type of network that we are filtering

	psqlDB := &PostgresDBService{
		ctx:              mainCtx,
		cancel:           cancel,
		connectionUrl:    url,
		psqlPool:         psqlPool,
		WriteChan:        make(chan WriteTask, WRITE_CHAN_LENGTH*workerNum),
		doneTasks:        make(chan interface{}, 1),
		endProcess:       0,
		FinishSignalChan: make(chan struct{}, 1),
		workerNum:        workerNum,
		maxBatchQueue:    batchLen,
		WgDBWriter:       sync.WaitGroup{},
	}
	// init the psql db
	err = psqlDB.init(ctx, psqlDB.psqlPool)
	if err != nil {
		return psqlDB, errors.Wrap(err, "error initializing the tables of the psqldb")
	}
	go psqlDB.runWriters()
	return psqlDB, err
}

func (p *PostgresDBService) init(ctx context.Context, pool *pgxpool.Pool) error {
	// create the tables
	err := p.createScoreMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	err = p.createBlockMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	err = p.createMissedBlockMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	err = p.createAttMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	err = p.createReorgMetricsTable(ctx, pool)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresDBService) DoneTasks() {
	atomic.AddInt32(&p.endProcess, int32(1))
	log.Infof("Received finish signal")
}

func (p *PostgresDBService) runWriters() {
	log.Info("Launching Beacon State Writers")
	log.Infof("Launching %d Beacon State Writers", p.workerNum)
	for i := 0; i < p.workerNum; i++ {
		p.WgDBWriter.Add(1)
		go func(dbWriterID int) {
			defer p.WgDBWriter.Done()
			wlogWriter := log.WithField("DBWriter", dbWriterID)
			writeBatch := pgx_v4.Batch{}
		loop:
			for {

				if p.endProcess >= 1 && len(p.WriteChan) == 0 {
					wlogWriter.Warnf("finish detected, closing persister")
					break loop
				}

				select {
				case task := <-p.WriteChan:
					wlogWriter.Tracef("Received new write task")
					writeBatch.Queue(task.QueryString, task.Params...)

					if writeBatch.Len() >= p.maxBatchQueue {
						wlogWriter.Tracef("Writing batch to database")
						err := p.ExecuteBatch(writeBatch)
						if err != nil {
							wlogWriter.Errorf("Error processing batch", err.Error())
						}
						writeBatch = pgx_v4.Batch{}
					} else {
						log.Tracef("%d pending tasks to persist", p.maxBatchQueue-writeBatch.Len())
					}

				case <-p.ctx.Done():
					wlogWriter.Info("shutdown detected, closing persister")
					break loop
				default:
				}

			}
			wlogWriter.Debugf("DB Writer finished...")

		}(i)
	}

	p.WgDBWriter.Wait()
	p.Close()

}

func (p PostgresDBService) Close() {
	p.psqlPool.Close()
}

type WriteTask struct {
	QueryString string
	Params      []interface{}
}

func (p PostgresDBService) ExecuteBatch(batch pgx_v4.Batch) error {

	snapshot := time.Now()
	tx, err := p.psqlPool.Begin(p.ctx)
	if err != nil {
		panic(err)
	}

	batchResults := tx.SendBatch(p.ctx, &batch)

	var qerr error
	var rows pgx_v4.Rows
	for qerr == nil {
		rows, qerr = batchResults.Query()
		rows.Close()
	}
	if qerr.Error() != "no result" {
		log.Errorf(qerr.Error())
	}

	log.Tracef("Batch process time: %f, batch size: %d", time.Since(snapshot).Seconds(), batch.Len())

	return tx.Commit(p.ctx)

}
