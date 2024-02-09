package analysis

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/attestantio/go-eth2-client/api"
)

func (b *ClientLiveData) PersistBlock(block api.VersionedProposal) error {
	if block.IsEmpty() {
		log.Errorf("attempt to persist empty block")
		return fmt.Errorf("empty block persist")
	}

	folder := fmt.Sprintf("%s/%s/%s/", b.blocksDir, b.Eth2Provider.Label, b.client)

	// check if foder exists

	if _, err := os.Stat(folder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			log.Errorf("could not create blocks dir: %s", err)
		}
	}

	slot, err := block.Slot()
	if err != nil {
		log.Errorf("could not persist block, slot not identified: %s", err)
		return err
	}

	// files are always new, block proposals are always in a new slot
	fileName := fmt.Sprintf("slot_%d.json", slot)
	fullPath := fmt.Sprintf("%s%s", folder, fileName)

	// create file
	f, err := os.Create(fullPath)
	if err != nil {
		log.Errorf("could not create file to persist block (%s): %s", fullPath, err)
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	written, err := w.WriteString(block.String())
	if err != nil {
		log.Errorf("error writing block to %s: %s", fullPath, err)
		return err
	}
	log.Debugf("wrote %d bytes to %s", written, fullPath)

	w.Flush()
	return nil
}
