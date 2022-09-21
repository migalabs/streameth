package score

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/client_api"
)

var (
	moduleName = "Scorer"
	log        = logrus.WithField(
		"module", moduleName)
)

type BlockScorer struct {
	ctx       context.Context
	BlockChan chan *client_api.APIBlockAnswer
}

func NewBlockScorer(ctx context.Context) *BlockScorer {
	return &BlockScorer{
		ctx:       ctx,
		BlockChan: make(chan *client_api.APIBlockAnswer, 10),
	}
}

func (s *BlockScorer) ListenBlocks() {
	for {

		select {
		case <-s.ctx.Done():
			log.Infof("context has died, closing block listener routine")
			close(s.BlockChan)
			return

		case blockTask, ok := <-s.BlockChan:
			if !ok {
				log.Errorf("could not receive new block task")
			} else {
				log.Infof("Analyzing block from %s!", blockTask.Label)
			}

		default:
		}

	}
}

type BeaconBlockMetrics struct {
	Score float64
}
