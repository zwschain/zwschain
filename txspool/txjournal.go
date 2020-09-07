/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package txspool

import (
	"io"
	"os"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/common/log"
	"path/filepath"
)

// errNoActiveJournal is returned if a transaction is attempted to be inserted
// into the journal, but no such file is currently open.
var errNoActiveJournal = errors.New("no active journal")
// devNull is a WriteCloser that just discards anything written into it. Its
// goal is to allow the transaction journal to write into a fake journal when
// loading transactions on startup without printing warnings due to no file
// being readt for write.
type devNull struct{}

func (*devNull) Write(p []byte) (n int, err error) { return len(p), nil }
func (*devNull) Close() error                      { return nil }

// txJournal is a rotating log of transactions with the aim of storing locally
// created transactions to allow non-executed ones to survive node restarts.
type txJournal struct {
	path   string         // Filesystem path to store the transactions at
	writer io.WriteCloser // Output stream to write new transactions into
}

// newTxJournal creates a new transaction journal to
func newTxJournal(path string) *txJournal {
	return &txJournal{
		path: path,
	}
}
func getFileSize(filename string) int64 {
	var size int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		size = f.Size()
		return nil
	})
	return size
}

// load parses a transaction journal dump from disk, loading its contents into
// the specified pool.
func (journal *txJournal) load(add func(*TxPoolTransaction) error) error {
	// Skip the parsing if the journal file doens't exist at all
	if _, err := os.Stat(journal.path); os.IsNotExist(err) {
		return nil
	}
	if getFileSize(journal.path) <= 0 {
		log.Infof("the transaction.rlp is empty file.")
		return nil
	}

	// Open the journal for loading any past transactions
	input, err := os.Open(journal.path)
	if err != nil {
		return err
	}
	defer input.Close()

	// Temporarily discard any journal additions (don't double add on load)
	journal.writer = new(devNull)
	defer func() { journal.writer = nil }()

	// Inject all transactions from the journal into the pool
	stream := rlp.NewStream(input, 0)
	total, dropped := 0, 0

	for {
		// Parse the next transaction and terminate on error
		tx := new(TxPoolTransaction)
		if err = stream.Decode(tx); err != nil {
			if err.Error() == errors.ErrEOF.Error() {
				err = nil
			} else {
				log.Infof("decode error:%s", err.Error())
			}
			break
		}
		// Import the transaction and bump the appropriate progress counters
		total++
		if tx.Tx != nil {
			if err = add(tx); err != nil {
				log.Debug("Failed to add journaled transaction", "err", err)
				dropped++
				continue
			}
		} else {
			log.Debug("journal decode tx failed. ", "error", "tx is nil.")
		}
	}
	log.Info("Loaded local transaction journal", "transactions", total, "dropped", dropped)

	return err
}

// insert adds the specified transaction to the local disk journal.
func (journal *txJournal) insert(tx *TxPoolTransaction) error {
	if journal.writer == nil {
		return errNoActiveJournal
	}
	if err := rlp.Encode(journal.writer, tx); err != nil {
		return err
	}
	return nil
}

// rotate regenerates the transaction journal based on the current contents of
// the transaction pool.
func (journal *txJournal) rotate(all map[common.Hash]*TxPoolTransaction) error {
	// Close the current journal (if any is open)
	if journal.writer != nil {
		if err := journal.writer.Close(); err != nil {
			return err
		}
		journal.writer = nil
	}
	// Generate a new journal with the contents of the current pool
	replacement, err := os.OpenFile(journal.path+".new", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	journaled := 0
	for _, tx := range all {
		if err = rlp.Encode(replacement, tx); err != nil {
			replacement.Close()
			return err
		}
		journaled += 1
	}
	replacement.Close()

	// Replace the live journal with the newly generated one
	if err = os.Rename(journal.path+".new", journal.path); err != nil {
		return err
	}
	sink, err := os.OpenFile(journal.path, os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	journal.writer = sink
	log.Debug("Regenerated local transaction journal", "transactions", journaled, "accounts", len(all))

	return nil
}

// close flushes the transaction journal contents to disk and closes the file.
func (journal *txJournal) close() error {
	var err error
	if journal.writer != nil {
		err = journal.writer.Close()
		journal.writer = nil
	}
	return err
}
