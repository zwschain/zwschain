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

package dag

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/types"
	dagcommon "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/memunit"
	"github.com/palletone/go-palletone/dag/migration"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/palletcache"
	"github.com/palletone/go-palletone/dag/storage"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/txspool"
)

type Dag struct {
	Cache                  palletcache.ICache
	Db                     ptndb.Database
	currentUnit            atomic.Value
	tokenEngine            tokenengine.ITokenEngine
	unstableUnitRep        dagcommon.IUnitRepository
	unstableUtxoRep        dagcommon.IUtxoRepository
	unstableStateRep       dagcommon.IStateRepository
	unstablePropRep        dagcommon.IPropRepository
	unstableUnitProduceRep dagcommon.IUnitProduceRepository

	stableUnitRep        dagcommon.IUnitRepository
	stableUtxoRep        dagcommon.IUtxoRepository
	stableStateRep       dagcommon.IStateRepository
	stablePropRep        dagcommon.IPropRepository
	stableUnitProduceRep dagcommon.IUnitProduceRepository

	ChainHeadFeed *event.Feed

	Mutex           sync.RWMutex
	Memdag          memunit.IMemDag                     // memory unit
	PartitionMemDag map[modules.AssetId]memunit.IMemDag //其他分区的MemDag

	//applyLock sync.Mutex

	//SPV
	rmLogsFeed event.Feed
	chainFeed  event.Feed

	//chainSideFeed event.Feed
	chainHeadFeed event.Feed
	logsFeed      event.Feed
	scope         event.SubscriptionScope

	unstableRepositoryUpdatedFeed  event.Feed
	unstableRepositoryUpdatedScope event.SubscriptionScope
}

var ContractChainId = "palletone"

func cache() palletcache.ICache {
	return freecache.NewCache(1000 * 1024)
}

func (d *Dag) IsEmpty() bool {
	it := d.Db.NewIterator()
	return !it.Next()
}

// return stable unit in dag
func (d *Dag) CurrentUnit(token modules.AssetId) *modules.Unit {
	hash, _, err := d.stablePropRep.GetNewestUnit(token)
	if err != nil {
		log.Errorf("Get stable unit by token[%s] error:%s", token.String(), err.Error())
		return nil
	}
	unit, err := d.GetUnitByHash(hash)
	if err != nil {
		return nil
	}

	return unit
}

func (d *Dag) GetStableChainIndex(token modules.AssetId) *modules.ChainIndex {
	memdag, err := d.getMemDag(token)
	if err != nil {
		log.Errorf("GetStableChainIndex by token[%s] error:%s", token.String(), err.Error())
		return nil
	}
	_, height := memdag.GetLastStableUnitInfo()
	return &modules.ChainIndex{AssetID: token, Index: height}
}

// return last main chain unit in memdag
func (d *Dag) GetMainCurrentUnit() *modules.Unit {
	return d.Memdag.GetLastMainChainUnit()
}

// return higher unit in memdag
func (d *Dag) GetCurrentUnit(assetId modules.AssetId) *modules.Unit {
	memUnit := d.GetCurrentMemUnit(assetId)
	curUnit := d.CurrentUnit(assetId)

	if curUnit == nil {
		log.Warnf("curUnit is nil.asset_id[%s], memunit[%s]", assetId.String(), memUnit.Hash().String())
		return memUnit
	}

	if memUnit == nil {
		log.Warnf("memUnit is nil.asset_id[%s], curunit[%s]", assetId.String(), curUnit.Hash().String())
		return curUnit
	}

	if curUnit.NumberU64() >= memUnit.NumberU64() {
		return curUnit
	}

	return memUnit
}

// return latest unit in the memdag of assetid
//func (d *Dag) GetCurrentMemUnit(assetId modules.AssetId, index uint64) *modules.Unit {
func (d *Dag) GetCurrentMemUnit(assetId modules.AssetId) *modules.Unit {
	//memdag, err := d.getMemDag(assetId)
	//if err != nil {
	//	log.Errorf("GetCurrentMemUnit by token[%s] error:%s", assetId.String(), err.Error())
	//	return nil
	//}
	//curUnit := memdag.GetLastMainChainUnit()

	hash, _, err := d.unstablePropRep.GetNewestUnit(assetId)
	if err != nil {
		log.Errorf("GetNewestUnit by token[%s] error:%s", assetId.String(), err.Error())
		return nil
	}

	curUnit, err := d.GetUnitByHash(hash)
	if err != nil {
		return nil
	}

	return curUnit
}

// return the unit exist in dag is true or false
func (d *Dag) HasUnit(hash common.Hash) bool {
	u, err := d.unstableUnitRep.GetUnit(hash)
	if err != nil {
		return false
	}
	return u != nil
}

// return the transaction exist in dag is true or false
func (d *Dag) IsTransactionExist(hash common.Hash) (bool, error) {
	return d.unstableUnitRep.IsTransactionExist(hash)
}

// return the unit confirmed is true or false
//func (d *Dag) UnitIsConfirmedByHash(hash common.Hash) bool {
//	if d.HasUnit(hash) {
//		return true
//	}
//	return false
//}

// return the unit's parent confirmed is true or false
func (d *Dag) ParentsIsConfirmByHash(hash common.Hash) bool {
	unit, err := d.GetUnitByHash(hash)
	if err != nil {
		return false
	}
	parents := unit.ParentHash()
	if len(parents) > 0 {
		par := parents[0]
		return d.HasUnit(par)
	}
	return false
}

// return the unit by chain index
func (d *Dag) GetUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error) {
	hash, err := d.unstableUnitRep.GetHashByNumber(number)
	if err != nil {
		log.Debug("GetUnitByNumber dagdb.GetHashByNumber err:", "error", err)
		return nil, err
	}
	return d.unstableUnitRep.GetUnit(hash)
}

// return all unstable units in memdag
func (d *Dag) GetUnstableUnits() []*modules.Unit {
	units := d.Memdag.GetChainUnits()
	result := modules.Units{}
	for _, u := range units {
		result = append(result, u)
	}
	sort.Sort(result)
	return result
}

// return the header by hash in dag
func (d *Dag) GetHeaderByHash(hash common.Hash) (*modules.Header, error) {
	//rep, _, _, _, _ := d.Memdag.GetUnstableRepositories()
	//uHeader, err := rep.GetHeaderByHash(hash)
	uHeader, err := d.unstableUnitRep.GetHeaderByHash(hash)

	if errors.IsNotFoundError(err) {
		uHeader, err = d.getHeaderByHashFromPMemDag(hash)
	}

	//if err != nil {
	//	uHeader, err = d.stableUnitRep.GetHeaderByHash(hash)
	//}

	if err != nil {
		return nil, err
	}

	return uHeader, nil
}
func (d *Dag) GetHeadersByAuthor(authorAddr common.Address, startHeight, count uint64) ([]*modules.Header, error) {
	return d.unstableUnitRep.GetHeadersByAuthor(authorAddr, startHeight, count)
}

// return the header by hash in memdag
func (d *Dag) getHeaderByHashFromPMemDag(hash common.Hash) (*modules.Header, error) {
	for _, memdag := range d.PartitionMemDag {
		h, e := memdag.GetHeaderByHash(hash)
		if e == nil {
			return h, e
		}
	}
	return nil, errors.ErrNotFound
}

// return the header by chain index in memdag
func (d *Dag) getHeaderByNumberFromPMemDag(number *modules.ChainIndex) (*modules.Header, error) {
	for _, memdag := range d.PartitionMemDag {
		h, e := memdag.GetHeaderByNumber(number)
		if e == nil {
			return h, e
		}
	}
	return nil, errors.ErrNotFound
}

// return the header by chain index in dag
func (d *Dag) GetHeaderByNumber(number *modules.ChainIndex) (*modules.Header, error) {
	uHeader, err := d.unstableUnitRep.GetHeaderByNumber(number)
	if errors.IsNotFoundError(err) {
		uHeader, err = d.getHeaderByNumberFromPMemDag(number)
	}
	if err != nil {
		log.Debug("GetHeaderByNumber failed ", "error:", err, "number", number.String())
		return nil, err
	}
	return uHeader, nil
}

// FastSyncCommitHead sets the current head block to the one defined by the hash
// irrelevant what the chain contents were prior.
func (d *Dag) FastSyncCommitHead(hash common.Hash) error {
	unit, err := d.GetUnitByHash(hash)
	if err != nil {
		return fmt.Errorf("non existent unit [%x...]", hash[:4])
	}
	// store current unit
	d.Mutex.Lock()
	d.currentUnit.Store(unit)
	d.Mutex.Unlock()

	return nil
}

// InsertDag attempts to insert the given batch of blocks in to the canonical
// chain or, otherwise, create a fork. If an error is returned it will return
// the index number of the failing block as well an error describing what went
// wrong.
// After insertion is done, all accumulated events will be fired.
// reference : Eth InsertChain
func (d *Dag) InsertDag(units modules.Units, txpool txspool.ITxPool, is_stable bool) (int, error) {
	count := int(0)

	for i, u := range units {
		// 重复验证 comment by albert
		//// all units must be continuous
		//if i > 0 && units[i].UnitHeader.Number.Index != units[i-1].UnitHeader.Number.Index+1 {
		//	return count, fmt.Errorf("Insert dag error: child height are not continuous, "+
		//		"parent unit number=%d, hash=%s; "+"child unit number=%d, hash=%s",
		//		units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash.String(),
		//		units[i].UnitHeader.Number.Index, units[i].UnitHash.String())
		//}
		//if i > 0 && !u.ContainsParent(units[i-1].UnitHash) {
		//	return count, fmt.Errorf("Insert dag error: child parents are not continuous, "+
		//		"parent unit number=%d, hash=%s; "+"child unit number=%d, hash=%s",
		//		units[i-1].UnitHeader.Number.Index, units[i-1].UnitHash.String(),
		//		units[i].UnitHeader.Number.Index, units[i].UnitHash.String())
		//}

		t1 := time.Now()
		timestamp := time.Unix(u.Timestamp(), 0)
		log.Debugf("Start InsertDag unit(%v) #%v parent(%v) @%v signed by %v", u.Hash().TerminalString(),
			u.NumberU64(), u.ParentHash()[0].TerminalString(), timestamp.Format("2006-01-02 15:04:05"),
			u.Author().Str())

		if is_stable {
			err := d.Memdag.AddStableUnit(u)
			if err != nil {
				log.Errorf("AddStableUnit failed,error:%s", err.Error())
				return i, err
			}
		} else {
			a, b, c, dd, e, err := d.Memdag.AddUnit(u, txpool, false)
			if err != nil {
				//return count, err
				log.Errorf("Memdag addUnit[%s] #%d signed by %v error:%s",
					u.Hash().String(), u.NumberU64(), u.Author().Str(), err.Error())
				return count, err
			} else if a != nil {
				if d.unstableUnitProduceRep != e {
					log.Debugf("send UnstableRepositoryUpdatedEvent")
					go d.unstableRepositoryUpdatedFeed.Send(modules.UnstableRepositoryUpdatedEvent{})
				}

				d.unstableUnitRep = a
				d.unstableUtxoRep = b
				d.unstableStateRep = c
				d.unstablePropRep = dd
				d.unstableUnitProduceRep = e
			}
		}

		log.Debugf("InsertDag[%s] #%d spent time:%s", u.Hash().String(), u.NumberU64(), time.Since(t1))

		if !is_stable && u.NumberU64()%1000 == 0 {
			log.Infof("Insert unit[%s] #%d to local", u.Hash().TerminalString(), u.NumberU64())
		}

		count += 1
	}

	if is_stable {
		tunitRep, tutxoRep, tstateRep, tpropRep, tUnitProduceRep := d.Memdag.GetUnstableRepositories()
		if tUnitProduceRep != d.unstableUnitProduceRep {
			log.Debugf("send UnstableRepositoryUpdatedEvent")
			go d.unstableRepositoryUpdatedFeed.Send(modules.UnstableRepositoryUpdatedEvent{})
		}

		d.unstableUnitRep = tunitRep
		d.unstableUtxoRep = tutxoRep
		d.unstableStateRep = tstateRep
		d.unstablePropRep = tpropRep
		d.unstableUnitProduceRep = tUnitProduceRep
	}

	return count, nil
}

// GetBlockHashesFromHash retrieves a number of block hashes starting at a given
// hash, fetching towards the genesis block.
func (d *Dag) GetUnitHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	header, err := d.GetHeaderByHash(hash)
	if err != nil {
		return nil
	}
	// Iterate the headers until enough is collected or the genesis reached
	chain := make([]common.Hash, 0, max)
	for i := uint64(0); i < max; i++ {
		if header.Index() == 0 {
			break
		}
		next := header.ParentHash()[0]
		h, err := d.unstableUnitRep.GetHeaderByHash(next)
		if err != nil {
			break
		}
		header = h
		chain = append(chain, next)
	}
	return chain
}

// return has header exist in dag or not by hash
func (d *Dag) HasHeader(hash common.Hash, number uint64) bool {
	h, _ := d.GetHeaderByHash(hash)
	return h != nil
}

// return has header exist in dag or not by hash
func (d *Dag) IsHeaderExist(hash common.Hash) bool {
	exist, _ := d.unstableUnitRep.IsHeaderExist(hash)
	return exist
}

// return latest header by assetId
func (d *Dag) CurrentHeader(token modules.AssetId) *modules.Header {
	memdag, err := d.getMemDag(token)
	if err != nil {
		log.Errorf("CurrentHeader by token[%s] error:%s", token.String(), err.Error())
		return nil
	}
	// 从memdag 获取最新的header
	unit := memdag.GetLastMainChainUnit()
	stable_hash, s_index := memdag.GetLastStableUnitInfo()
	if unit != nil {
		if unit.NumberU64() >= s_index {
			return unit.Header()
		}
	}
	if stable_header, err := d.GetHeaderByHash(stable_hash); err != nil {
		return nil
	} else {
		return stable_header
	}
}

// return unit's body , all transactions of unit by hash
func (d *Dag) GetUnitTransactions(hash common.Hash) (modules.Transactions, error) {
	return d.unstableUnitRep.GetUnitTransactions(hash)
}

// GetUnitTxsHash is return the unit's txs hash list.
func (d *Dag) GetUnitTxsHash(hash common.Hash) ([]common.Hash, error) {
	return d.unstableUnitRep.GetBody(hash)
}

// return the transaction with unit packed by hash
func (d *Dag) GetTransaction(hash common.Hash) (*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetTransaction(hash)
}

// return the transaction with unit info by reqid
func (d *Dag) GetTxByReqId(reqid common.Hash) (*modules.TransactionWithUnitInfo, error) {
	hash, err := d.unstableUnitRep.GetTxHashByReqId(reqid)
	if err != nil {
		return nil, err
	}
	return d.unstableUnitRep.GetTransaction(hash)
}

// return the transaction by hash
func (d *Dag) GetTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return d.unstableUnitRep.GetTransactionOnly(hash)
}

// return the stable transaction by hash
func (d *Dag) GetStableTransactionOnly(hash common.Hash) (*modules.Transaction, error) {
	return d.stableUnitRep.GetTransactionOnly(hash)
}

// return the stable unit by hash or number
func (d *Dag) GetStableUnit(hash common.Hash) (*modules.Unit, error) {
	return d.stableUnitRep.GetUnit(hash)
}

// return the stable unit by chain index
func (d *Dag) GetStableUnitByNumber(number *modules.ChainIndex) (*modules.Unit, error) {
	hash, err := d.stableUnitRep.GetHashByNumber(number)
	if err != nil {
		log.Debug("GetStableUnitByNumber dagdb.GetHashByNumber err:", "error", err)
		return nil, err
	}
	return d.stableUnitRep.GetUnit(hash)
}

// retunr the txLookEntry by transaction hash
func (d *Dag) GetTxSearchEntry(hash common.Hash) (*modules.TxLookupEntry, error) {
	txlookup, err := d.unstableUnitRep.GetTxLookupEntry(hash)
	return txlookup, err
}

// InsertHeaderDag attempts to insert the given header chain in to the local
// chain, possibly creating a reorg. If an error is returned, it will return the
// index number of the failing header as well an error describing what went wrong.
//
// The verify parameter can be used to fine tune whether nonce verification
// should be done or not. The reason behind the optional check is because some
// of the header retrieval mechanisms already need to verify nonces, as well as
// because nonces can be verified sparsely, not needing to check each.
func (d *Dag) InsertHeaderDag(headers []*modules.Header) (int, error) {
	for i, header := range headers {
		err := d.saveHeader(header)

		if err != nil {
			return i, fmt.Errorf("InsertHeaderDag, on header:%d, at saveHeader Error:%s", i, err.Error())
		}
		if header.NumberU64()%2000 == 0 {
			log.Infof("Insert header[%s] #%d", header.Hash().String(), header.NumberU64())
		}
	}
	return len(headers), nil
}

// refresh partition memdag when newdag or system contract state be changed.
func (d *Dag) RefreshPartitionMemDag() {
	db := d.Db
	unitRep := d.stableUnitRep
	propRep := d.stablePropRep

	mainChain, err := d.stableStateRep.GetMainChain()
	if err == nil && mainChain != nil {
		//分区上要通过Memdag保留PTN的Header
		if d.PartitionMemDag == nil {
			d.PartitionMemDag = make(map[modules.AssetId]memunit.IMemDag)
		}
		mainChainMemDag, ok := d.PartitionMemDag[mainChain.GasToken]
		threshold := int(mainChain.StableThreshold)
		if !ok {
			d.initDataForMainChainHeader(mainChain)
			log.Debugf("Init main chain mem dag for:%s", mainChain.GasToken.String())
			pmemdag := memunit.NewMemDag(mainChain.GasToken, threshold, true, db, unitRep,
				propRep, d.stableStateRep, cache(), d.tokenEngine)
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep,
			// d.unstablePropRep, d.unstableUnitProduceRep)
			d.PartitionMemDag[mainChain.GasToken] = pmemdag
		} else {
			mainChainMemDag.SetStableThreshold(threshold) //可能更新了该数字
		}
	} else {
		log.Info("Don't have main chain config for partition")
	}
	partitions, err := d.stableStateRep.GetPartitionChains()
	if err != nil {
		log.Warnf("GetPartitionChains error:%s", err.Error())
		return
	}
	log.Debug("Start to refresh partition mem dag")
	//Init partition memdag
	if d.PartitionMemDag == nil {
		partitionMemdag := make(map[modules.AssetId]memunit.IMemDag)

		for _, partition := range partitions {
			ptoken := partition.GasToken
			threshold := int(partition.StableThreshold)
			d.initDataForPartition(partition)
			log.Debugf("Init partition mem dag for:%s", ptoken.String())
			pmemdag := memunit.NewMemDag(ptoken, threshold, true, db, unitRep, propRep,
				d.stableStateRep, cache(), d.tokenEngine)
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep,
			// d.unstablePropRep, d.unstableUnitProduceRep)
			partitionMemdag[ptoken] = pmemdag
		}

		d.PartitionMemDag = partitionMemdag
		return
	}
	//Exist! update
	for _, partition := range partitions {
		ptoken := partition.GasToken
		threshold := int(partition.StableThreshold)
		partitonMemDag, ok := d.PartitionMemDag[ptoken]
		if !ok {
			d.initDataForPartition(partition)
			log.Debugf("Init partition mem dag for:%s", ptoken.String())
			pmemdag := memunit.NewMemDag(ptoken, threshold, true, db, unitRep, propRep,
				d.stableStateRep, cache(), d.tokenEngine)
			//pmemdag.SetUnstableRepositories(d.unstableUnitRep, d.unstableUtxoRep, d.unstableStateRep,
			// d.unstablePropRep, d.unstableUnitProduceRep)
			d.PartitionMemDag[ptoken] = pmemdag
		} else {
			partitonMemDag.SetStableThreshold(threshold) //可能更新了该数字
		}
	}

}

// init the partition's data
func (d *Dag) initDataForPartition(partition *modules.PartitionChain) {
	pHeader := partition.GetGenesisHeader()
	exist, _ := d.stableUnitRep.IsHeaderExist(pHeader.Hash())
	if !exist {
		log.Debugf("Init partition[%s] genesis header:%s",
			pHeader.ChainIndex().AssetID.String(), pHeader.Hash().String())
		d.stableUnitRep.SaveNewestHeader(pHeader)
	}
}

// init data for main chain header
func (d *Dag) initDataForMainChainHeader(mainChain *modules.MainChain) {
	pHeader := mainChain.GetGenesisHeader()
	exist, _ := d.stableUnitRep.IsHeaderExist(pHeader.Hash())
	if !exist {
		log.Debugf("Init main chain[%s] genesis header:%s",
			pHeader.ChainIndex().AssetID.String(), pHeader.Hash().String())
		d.stableUnitRep.SaveNewestHeader(pHeader)
	}
}

// newDag, with db , light to build a new dag
// firstly to check db migration, is updated ptn database.
func NewDag(db ptndb.Database, cache palletcache.ICache, light bool) (*Dag, error) {
	tokenEngine := tokenengine.Instance //TODO Devin tokenENgine from parmeter
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenEngine)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	err := checkDbMigration(db, stateDb)
	if err != nil {
		return nil, err
	}

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb, tokenEngine)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenEngine)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb, dagDb)
	stableUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)
	gasToken := dagconfig.DagConfig.GetGasToken()
	threshold, _ := propRep.GetChainThreshold()
	unstableChain := memunit.NewMemDag(gasToken, threshold, light, db,
		unitRep, propRep, stateRep, cache, tokenEngine)
	tunitRep, tutxoRep, tstateRep, tpropRep, tUnitProduceRep := unstableChain.GetUnstableRepositories()

	dag := &Dag{
		Db:                     db,
		tokenEngine:            tokenEngine,
		Cache:                  cache,
		unstableUnitRep:        tunitRep,
		unstableUtxoRep:        tutxoRep,
		unstableStateRep:       tstateRep,
		unstablePropRep:        tpropRep,
		unstableUnitProduceRep: tUnitProduceRep,
		stablePropRep:          propRep,
		stableUnitRep:          unitRep,
		stableUtxoRep:          utxoRep,
		stableStateRep:         stateRep,
		stableUnitProduceRep:   stableUnitProduceRep,
		ChainHeadFeed:          new(event.Feed),
		Memdag:                 unstableChain,
	}
	dag.stableUnitRep.SubscribeSysContractStateChangeEvent(dag.AfterSysContractStateChangeEvent)
	dag.stableUnitProduceRep.SubscribeChainMaintenanceEvent(dag.AfterChainMaintenanceEvent)
	dag.Memdag.SubscribeSwitchMainChainEvent(dag.SwitchMainChainEvent)
	//hash, chainIndex, _ := dag.stablePropRep.GetNewestUnit(gasToken)
	log.Infof("newDag success, current unit, chain info[%s]", gasToken.String())
	// init partition memdag
	dag.RefreshPartitionMemDag()
	return dag, nil
}

// check db migration ,to upgrade ptn database
func checkDbMigration(db ptndb.Database, stateDb storage.IStateDb) error {
	//特殊处理
	//  获取陪审员列表
	//updateStateDbJurys(db)

	// 获取旧的gptn版本号
	old_vertion, err := stateDb.GetDataVersion()
	if err != nil {
		log.Warn("Don't have database version, Ignore data migration")
		return nil
	}
	log.Debugf("the database version is:%s", old_vertion.Version)
	//fmt.Printf("the database version is:%s\n", old_vertion.Version)

	// 获取当前gptn版本号
	now_version := configure.Version
	log.Debugf("the program version is:%s", now_version)
	//fmt.Printf("the program version is:%s\n", now_version)
	next_version := old_vertion.Version

	if next_version != now_version {
		t := time.Now()
		log.Infof("Start migration,upgrade gtpn vertion[%s] to [%s], it may spend a long time, please wait...",
			next_version, now_version)
		// migrations
		mig_versions := migration.NewMigrations(db)

		for {
			if mig, has := mig_versions[next_version]; has {
				if err := mig.ExecuteUpgrade(); err != nil {
					return err
				}
				next_version = mig.ToVersion()
				data_version := new(modules.DataVersion)
				data_version.Name = "gptn"
				data_version.Version = next_version
				stateDb.SaveDataVersion(data_version)
			}

			if next_version == now_version {
				break
			}

			//// 版本升级超时处理
			//if now := time.Now(); now.After(t.Add(1 * time.Minute)) {
			//	log.Infof("upgrade gptn failed. error: timeout[%s]", time.Since(t))
			//	break
			//}
		}

		log.Infof("Complete migration, spent time:%s", time.Since(t))
		return nil
	}

	return nil
}

// to refresh partition memdag
func (dag *Dag) AfterSysContractStateChangeEvent(arg *modules.SysContractStateChangeEvent) {
	log.Debug("Process AfterSysContractStateChangeEvent")
	if bytes.Equal(arg.ContractId, syscontract.PartitionContractAddress.Bytes()) {
		//分区合约进行了修改，刷新PartitionMemDag
		dag.RefreshPartitionMemDag()
	}
}

// after chain maintenance to set stable threshold
func (dag *Dag) AfterChainMaintenanceEvent(arg *modules.ChainMaintenanceEvent) {
	log.Debug("Process AfterChainMaintenanceEvent")
	//换届完成，dag需要进行的操作：
	threshold, _ := dag.unstablePropRep.GetChainThreshold()
	dag.Memdag.SetStableThreshold(threshold)
}

func (dag *Dag) SwitchMainChainEvent(arg *memunit.SwitchMainChainEvent) {
	log.Debugf("Switch main chain event!!! old unit:%s %d,new unit:%s %d", arg.OldLastUnit.Hash().TerminalString(),
		arg.OldLastUnit.NumberU64(), arg.NewLastUnit.Hash().TerminalString(), arg.NewLastUnit.NumberU64())
}

// to build a new dag when init genesis
func NewDag4GenesisInit(db ptndb.Database) (*Dag, error) {
	tokenEngine := tokenengine.Instance
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenEngine)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb, tokenEngine)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenEngine)
	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb, dagDb)

	statleUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)

	dag := &Dag{
		Db:                   db,
		tokenEngine:          tokenEngine,
		stableUnitRep:        unitRep,
		stableUtxoRep:        utxoRep,
		stablePropRep:        propRep,
		stableStateRep:       stateRep,
		stableUnitProduceRep: statleUnitProduceRep,
		ChainHeadFeed:        new(event.Feed),
		unstableUnitRep:      unitRep,
		unstablePropRep:      propRep,
		unstableStateRep:     stateRep,
		unstableUtxoRep:      utxoRep,
	}
	return dag, nil
}

// to build a dag for test
func NewDagForTest(db ptndb.Database) (*Dag, error) {
	tokenEngine := tokenengine.Instance
	dagDb := storage.NewDagDb(db)
	utxoDb := storage.NewUtxoDb(db, tokenEngine)
	stateDb := storage.NewStateDb(db)
	idxDb := storage.NewIndexDb(db)
	propDb := storage.NewPropertyDb(db)

	propRep := dagcommon.NewPropRepository(propDb)
	stateRep := dagcommon.NewStateRepository(stateDb, dagDb)
	utxoRep := dagcommon.NewUtxoRepository(utxoDb, idxDb, stateDb, propDb, tokenEngine)
	unitRep := dagcommon.NewUnitRepository(dagDb, idxDb, utxoDb, stateDb, propDb, tokenEngine)
	statleUnitProduceRep := dagcommon.NewUnitProduceRepository(unitRep, propRep, stateRep)

	threshold, _ := propRep.GetChainThreshold()
	unstableChain := memunit.NewMemDag(modules.PTNCOIN, threshold, false, db, unitRep,
		propRep, stateRep, cache(), tokenEngine)
	tunitRep, tutxoRep, tstateRep, tpropRep, tUnitProduceRep := unstableChain.GetUnstableRepositories()

	dag := &Dag{
		Db:                     db,
		tokenEngine:            tokenEngine,
		stableUnitRep:          unitRep,
		stableUtxoRep:          utxoRep,
		stableStateRep:         stateRep,
		stablePropRep:          propRep,
		stableUnitProduceRep:   statleUnitProduceRep,
		ChainHeadFeed:          new(event.Feed),
		Memdag:                 unstableChain,
		unstableUnitRep:        tunitRep,
		unstableUtxoRep:        tutxoRep,
		unstableStateRep:       tstateRep,
		unstablePropRep:        tpropRep,
		unstableUnitProduceRep: tUnitProduceRep,
	}
	return dag, nil
}

// get contract by contract id
func (d *Dag) GetContract(id []byte) (*modules.Contract, error) {
	return d.unstableStateRep.GetContract(id)
}

func (d *Dag) SaveContract(contract *modules.Contract) error {
	return d.unstableStateRep.SaveContract(contract)
}

// get contract deploy by tempId, contractId, and name
func (d *Dag) GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error) {
	return d.unstableStateRep.GetContractDeploy(tempId, contractId, name)
}

// get unit chain index by hash
func (d *Dag) GetUnitNumber(hash common.Hash) (*modules.ChainIndex, error) {
	return d.unstableUnitRep.GetNumberWithUnitHash(hash)
}

// get trie sync progress
func (d *Dag) GetTrieSyncProgress() (uint64, error) {
	return d.unstableUnitRep.GetTrieSyncProgress()
}

// get the utxoEntry by outpoint
func (d *Dag) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.unstableUtxoRep.GetUtxoEntry(outpoint)
}

// get the stxoEntry by outpoint
func (d *Dag) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	return d.unstableUtxoRep.GetStxoEntry(outpoint)
}

// get the txoutput by outpoint include UTXO and STXO
func (d *Dag) GetTxOutput(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	utxo, err := d.unstableUtxoRep.GetUtxoEntry(outpoint)
	if err == nil {
		return utxo, err
	}
	stxo, err := d.unstableUtxoRep.GetStxoEntry(outpoint)
	if err != nil {
		return nil, err
	}
	u := &modules.Utxo{
		Amount:   stxo.Amount,
		Asset:    stxo.Asset,
		PkScript: stxo.PkScript,
		LockTime: stxo.LockTime,
	}
	return u, nil
}

// get the tx's  utxoView
func (d *Dag) GetUtxoView(tx *modules.Transaction) (*txspool.UtxoViewpoint, error) {
	neededSet := make(map[modules.OutPoint]struct{})

	for _, msgcopy := range tx.TxMessages() {
		if msgcopy.App == modules.APP_PAYMENT {
			if msg, ok := msgcopy.Payload.(*modules.PaymentPayload); ok {
				if !msg.IsCoinbase() {
					for _, in := range msg.Inputs {
						neededSet[*in.PreviousOutPoint] = struct{}{}
					}
				}
			}
		}
	}

	view := txspool.NewUtxoViewpoint()
	d.Mutex.RLock()
	defer d.Mutex.RUnlock()
	err := view.FetchUtxos(d.unstableUtxoRep, neededSet)
	return view, err
}

// get the tx's utxoViewpoint
func (d *Dag) GetUtxosOutViewbyTx(tx *modules.Transaction) *txspool.UtxoViewpoint {
	view := txspool.NewUtxoViewpoint()
	view.AddTxOuts(tx)
	return view
}

// get the unit's utxoViewPoint
func (d *Dag) GetUtxosOutViewbyUnit(unit *modules.Unit) *txspool.UtxoViewpoint {
	txs := unit.Transactions()
	view := txspool.NewUtxoViewpoint()
	for _, tx := range txs {
		vi := d.GetUtxosOutViewbyTx(tx)
		for key, utxo := range vi.Entries() {
			view.AddUtxo(key, utxo)
		}
	}
	return view
}

// return the true or false , is utxo is spent.
func (d *Dag) IsUtxoSpent(outpoint *modules.OutPoint) (bool, error) {
	return d.unstableUtxoRep.IsUtxoSpent(outpoint)
}

// return all utxo in dag
func (d *Dag) GetAllUtxos() (map[modules.OutPoint]*modules.Utxo, error) {
	d.Mutex.RLock()
	items, err := d.unstableUtxoRep.GetAllUtxos()
	d.Mutex.RUnlock()

	return items, err
}

// return all outpoint by address
func (d *Dag) GetAddrOutpoints(addr common.Address) ([]modules.OutPoint, error) {
	all, err := d.unstableUtxoRep.GetAddrOutpoints(addr)

	return all, err
}

// return address by outpoint
func (d *Dag) GetAddrByOutPoint(outPoint *modules.OutPoint) (common.Address, error) {
	utxo, err := d.unstableUtxoRep.GetUtxoEntry(outPoint)
	if err != nil {
		return common.Address{}, err
	}
	return d.tokenEngine.GetAddressFromScript(utxo.PkScript)
}

// return the transaction's fee ,
func (d *Dag) GetTxFee(pay *modules.Transaction) (*modules.AmountAsset, error) {
	return d.unstableUtxoRep.ComputeTxFee(pay)
}

// return all address by the transaction
func (d *Dag) GetTxFromAddress(tx *modules.Transaction) ([]common.Address, error) {
	return d.unstableUnitRep.GetTxFromAddress(tx)
}

// return all transaction with unit info by asset
func (d *Dag) GetAssetTxHistory(asset *modules.Asset) ([]*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetAssetTxHistory(asset)
}

// get the token balance by address and asset
func (d *Dag) GetAddr1TokenUtxos(addr common.Address, asset *modules.Asset) (
	map[modules.OutPoint]*modules.Utxo, error) {
	all, err := d.unstableUtxoRep.GetAddrUtxos(addr, asset)
	return all, err
}

// get all utxos in dag by address
func (d *Dag) GetAddrUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	all, err := d.unstableUtxoRep.GetAddrUtxos(addr, nil)

	return all, err
}
func (d *Dag) GetAddrStableUtxos(addr common.Address) (map[modules.OutPoint]*modules.Utxo, error) {
	all, err := d.stableUtxoRep.GetAddrUtxos(addr, nil)

	return all, err
}

// refresh system parameters
func (d *Dag) RefreshSysParameters() {
	d.unstableUnitProduceRep.RefreshSysParameters()
}

// return all transaction with unit info by address
func (d *Dag) GetAddrTransactions(addr common.Address) ([]*modules.TransactionWithUnitInfo, error) {
	return d.unstableUnitRep.GetAddrTransactions(addr)
}

// get contract state return codes, state version by contractId and field
func (d *Dag) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return d.unstableStateRep.GetContractState(id, field)
}
func (d *Dag) GetContractStateByVersion(id []byte, field string, version *modules.StateVersion) ([]byte, error) {
	return d.unstableStateRep.GetContractStateByVersion(id, field, version)
}

// get contract all state
func (d *Dag) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesById(id)
}

// return contract state value by contractId, prefix
func (d *Dag) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return d.unstableStateRep.GetContractStatesByPrefix(id, prefix)
}

// return electionInfo by contractId
func (d *Dag) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	return d.unstableStateRep.GetContractJury(contractId)
}

// save header
func (d *Dag) saveHeader(header *modules.Header) error {
	if header == nil {
		return errors.ErrNullPoint
	}
	asset := header.GetNumber().AssetID
	memdag, err := d.getMemDag(asset)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if err := memdag.SaveHeader(header); err != nil {
		return fmt.Errorf("Save Header in MemDag, occurred error: %s", err.Error())
	}

	return nil
}

// return a memdag by assetId
func (d *Dag) getMemDag(asset modules.AssetId) (memunit.IMemDag, error) {
	var memdag memunit.IMemDag
	gasToken := dagconfig.DagConfig.GetGasToken()
	if asset == gasToken {
		memdag = d.Memdag
	} else {
		memdag = d.PartitionMemDag[asset]
		if memdag == nil {
			return nil, errors.New("Don't have partition mem dag for token:" + asset.String())
		}
	}
	return memdag, nil
}

// save unit, 目前只用来存创世unit
func (d *Dag) SaveUnit(unit *modules.Unit, txpool txspool.ITxPool, isGenesis bool) error {
	// step1. check exists
	if !isGenesis {
		if d.IsHeaderExist(unit.Hash()) {
			return errors.ErrUnitExist
		}
	}
	if isGenesis {
		err := d.stableUnitRep.SaveUnit(unit, true)
		if err != nil {
			return err
		}
		// set memdag state
		return nil
	}

	if a, b, c, dd, e, err := d.Memdag.AddUnit(unit, txpool, false); err != nil {
		return fmt.Errorf("Save MemDag, occurred error: %s", err.Error())
	} else {
		if a != nil {
			if d.unstableUnitProduceRep != e {
				log.Debugf("send UnstableRepositoryUpdatedEvent")
				go d.unstableRepositoryUpdatedFeed.Send(modules.UnstableRepositoryUpdatedEvent{})
			}

			d.unstableUnitRep = a
			d.unstableUtxoRep = b
			d.unstableStateRep = c
			d.unstablePropRep = dd
			d.unstableUnitProduceRep = e
		}
	}

	return nil
}

// return genesis unit of ptn chain
func (d *Dag) GetGenesisUnit() (*modules.Unit, error) {
	return d.stableUnitRep.GetGenesisUnit()
}

// return contract template by tlpId
func (d *Dag) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return d.unstableStateRep.GetContractTpl(tplId)
}

// return all  contract templates
func (d *Dag) GetAllContractTpl() ([]*modules.ContractTemplate, error) {
	return d.unstableStateRep.GetAllContractTpl()
}

// return the tplid's tlp code
func (d *Dag) GetContractTplCode(tplId []byte) ([]byte, error) {
	return d.unstableStateRep.GetContractTplCode(tplId)
}

// return the chain index by assetId
func (d *Dag) GetCurrentUnitIndex(token modules.AssetId) (*modules.ChainIndex, error) {
	currentUnit := d.GetCurrentUnit(token)
	//	return d.GetUnitNumber(currentUnitHash)
	return currentUnit.Number(), nil
}

// dag's common geter, return the key's value
func (d *Dag) GetCommon(key []byte, stableDb bool) ([]byte, error) {
	if stableDb {
		return d.stableUnitRep.GetCommon(key)
	}
	return d.unstableUnitRep.GetCommon(key)
}

// return the prefix's all key && value.
func (d *Dag) GetCommonByPrefix(prefix []byte, stableDb bool) map[string][]byte {
	if stableDb {
		return d.stableUnitRep.GetCommonByPrefix(prefix)
	}
	return d.unstableUnitRep.GetCommonByPrefix(prefix)
}
func (d *Dag) GetAllData() ([][]byte, [][]byte) {
	return d.stableUnitRep.GetAllData()
}

// save the key, value
func (d *Dag) SaveCommon(key, val []byte) error {
	return d.stableUnitRep.SaveCommon(key, val)
}

// set the unit's group sign ,and set to be stable unit  by hash
func (d *Dag) SetUnitGroupSign(unitHash common.Hash, groupSign []byte, txpool txspool.ITxPool) error {
	if groupSign == nil {
		err := fmt.Errorf("this unit(%v)'s group sign is null", unitHash.TerminalString())
		log.Debug(err.Error())
		return err
	}

	log.Debugf("receive unit(%v)'s group sign: %v", unitHash.TerminalString(), hexutil.Encode(groupSign))

	// 判断本节点是否正在同步数据
	if !d.IsSynced(false) {
		err := "this node is synced"
		log.Debugf(err)
		return fmt.Errorf(err)
	}

	isStable, err := d.IsIrreversibleUnit(unitHash)
	if err != nil {
		return err
	}

	if isStable {
		// 或者由于网络延迟，该单元在收到群签名之前，已经根据深度转为不可逆了
		log.Debugf("this unit(%v) is already irreversible", unitHash.TerminalString())
		return nil
	}

	// 验证群签名：
	err = d.VerifyUnitGroupSign(unitHash, groupSign)
	if err != nil {
		return err
	}

	// 群签之后， 更新memdag，将该unit和它的父单元们稳定存储。
	d.Memdag.SetUnitGroupSign(unitHash, groupSign, txpool)

	//TODO albert 待合并
	// 状态更新
	//go d.updateGlobalPropDependGroupSign(unitHash)
	return nil
}

func (d *Dag) QueryDbByKey(key []byte) ([]byte, error) {
	return d.Db.Get(key)
}

func (d *Dag) QueryDbByPrefix(prefix []byte) ([]*modules.DbRow, error) {

	iter := d.Db.NewIteratorWithPrefix(prefix)
	var result []*modules.DbRow
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		result = append(result, &modules.DbRow{Key: key, Value: value})
	}
	return result, nil
}

// return a transaction hash by the reqId
func (d *Dag) GetTxHashByReqId(reqid common.Hash) (common.Hash, error) {
	return d.unstableUnitRep.GetTxHashByReqId(reqid)
}

// return a file info by the filehash
func (d *Dag) GetFileInfo(filehash []byte) ([]*modules.FileInfo, error) {
	return d.unstableUnitRep.GetFileInfo(filehash)
}

// Light Palletone Subprotocal
func (d *Dag) GetLightHeaderByHash(headerHash common.Hash) (*modules.Header, error) {
	return nil, nil
}

// return a light chain's height by the assetId
func (d *Dag) GetLightChainHeight(assetId modules.AssetId) uint64 {
	header := d.CurrentHeader(assetId)
	if header != nil {
		return header.GetNumber().Index
	}
	return uint64(0)
}

// insert headers into a light
func (d *Dag) InsertLightHeader(headers []*modules.Header) (int, error) {
	log.Debug("Dag InsertLightHeader numbers", "", len(headers))
	for _, header := range headers {
		log.Debug("Dag InsertLightHeader info", "header index:", header.Index(),
			"assetid", header.GetNumber().AssetID)
	}
	count, err := d.InsertHeaderDag(headers)

	return count, err
}

// All leaf nodes for dag downloader.
// 根据资产Id返回所有链的header。
func (d *Dag) GetAllLeafNodes() ([]*modules.Header, error) {
	// step1: get all AssetId
	partitions, _ := d.unstableStateRep.GetPartitionChains()
	var leafs []*modules.Header
	for _, partition := range partitions {
		tokenId := partition.GasToken
		pMemdag, ok := d.PartitionMemDag[tokenId]
		if ok {
			unit := pMemdag.GetLastMainChainUnit()
			leafs = append(leafs, unit.UnitHeader)
		} else {
			log.Warnf("Token[%s] is a patition, but not in dag.PartitionMemDag", tokenId.String())
		}
	}
	return leafs, nil
}

//SPV
// SubscribeRemovedLogsEvent registers a subscription of RemovedLogsEvent.
func (bc *Dag) SubscribeRemovedLogsEvent(ch chan<- modules.RemovedLogsEvent) event.Subscription {
	return bc.scope.Track(bc.rmLogsFeed.Subscribe(ch))
}

// SubscribeChainHeadEvent registers a subscription of ChainHeadEvent.
func (bc *Dag) SubscribeChainHeadEvent(ch chan<- modules.ChainHeadEvent) event.Subscription {
	return bc.scope.Track(bc.chainHeadFeed.Subscribe(ch))
}

// SubscribeLogsEvent registers a subscription of []*types.Log.
func (bc *Dag) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return bc.scope.Track(bc.logsFeed.Subscribe(ch))
}

// SubscribeChainEvent registers a subscription of ChainEvent.
func (bc *Dag) SubscribeChainEvent(ch chan<- modules.ChainEvent) event.Subscription {
	return bc.scope.Track(bc.chainFeed.Subscribe(ch))
}

// PostChainEvents iterates over the events generated by a chain insertion and
// posts them into the event feed.
// TODO: Should not expose PostChainEvents. The chain events should be posted in WriteBlock.
func (bc *Dag) PostChainEvents(events []interface{}) {
	log.Debug("enter PostChainEvents")

	for _, event := range events {
		switch ev := event.(type) {
		case modules.ChainEvent:
			bc.chainFeed.Send(ev)

		case modules.ChainHeadEvent:
			bc.chainHeadFeed.Send(ev)

			//case modules.ChainSideEvent:
			//	bc.chainSideFeed.Send(ev)
		}
	}
}

// return all partition chain from dag
func (bc *Dag) GetPartitionChains() ([]*modules.PartitionChain, error) {
	return bc.unstableStateRep.GetPartitionChains()
}

// return main chain from dag
func (bc *Dag) GetMainChain() (*modules.MainChain, error) {
	return bc.unstableStateRep.GetMainChain()
}

// return requester's address from the transaction
func (d *Dag) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	return d.unstableUnitRep.GetTxRequesterAddress(tx)
}

// refresh all address transaction's index
func (d *Dag) RefreshAddrTxIndex() error {
	return d.stableUnitRep.RefreshAddrTxIndex()
}

// return all contract from dag
func (d *Dag) GetAllContracts() ([]*modules.Contract, error) {
	return d.unstableStateRep.GetAllContracts()
}

// return all contract by tplId from dag
func (d *Dag) GetContractsByTpl(tplId []byte) ([]*modules.Contract, error) {
	return d.unstableStateRep.GetContractsByTpl(tplId)
}

// close a dag
func (d *Dag) Close() {
	d.unstableUnitProduceRep.Close()
	d.Memdag.Close()

	d.scope.Close()
	d.unstableRepositoryUpdatedScope.Close()

	for _, pmg := range d.PartitionMemDag {
		pmg.Close()
	}

	d.Db.Close()
	log.Debug("Close all dag database connections")
}

// store a data version in dag
func (d *Dag) StoreDataVersion(dv *modules.DataVersion) error {
	return d.stableStateRep.StoreDataVersion(dv)
}

// return a data version from dag
func (d *Dag) GetDataVersion() (*modules.DataVersion, error) {
	return d.stableStateRep.GetDataVersion()
}

// return proof of existence
func (d *Dag) QueryProofOfExistenceByReference(ref []byte) ([]*modules.ProofOfExistence, error) {
	return d.unstableUnitRep.QueryProofOfExistenceByReference(ref)
}

// return proof of existence by asset
func (d *Dag) GetAssetReference(asset []byte) ([]*modules.ProofOfExistence, error) {
	return d.unstableUnitRep.GetAssetReference(asset)
}

func (d *Dag) CheckReadSetValid(contractId []byte, readSet []modules.ContractReadSet) bool {
	return d.unstableUnitRep.CheckReadSetValid(contractId, readSet)
}

func (d *Dag) CheckHeaderCorrect(number int) error {
	ptn := modules.PTNCOIN
	if number == 0 {
		newestUnitHash, newestIndex, _ := d.stablePropRep.GetNewestUnit(ptn)
		log.Infof("Newest unit[%s] height:%d", newestUnitHash.String(), newestIndex.Index)
		number = int(newestIndex.Index)
	}
	header, err := d.stableUnitRep.GetHeaderByNumber(modules.NewChainIndex(ptn, uint64(number)))
	if err != nil {
		return fmt.Errorf("Unit height:%d not exits", number)
	}
	parentHash := header.ParentHash()[0]
	parentNumber := header.NumberU64() - 1
	for {
		header, err = d.stableUnitRep.GetHeaderByHash(parentHash)
		if err != nil {
			return fmt.Errorf("Unit :%s not exits", parentHash.String())
		}
		if header.NumberU64() != parentNumber {
			return fmt.Errorf("Number not correct,%d,%d", header.NumberU64(), parentNumber)
		}
		if len(header.ParentHash()) > 0 {
			parentHash = header.ParentHash()[0]
			parentNumber = header.NumberU64() - 1
		} else {
			log.Infof("Check complete!%d", header.NumberU64())
			break
		}
		if header.NumberU64()%1000 == 0 {
			log.Infof("Check header correct:%d", header.NumberU64())
		}
	}
	return nil
}
func (d *Dag) CheckUnitsCorrect(assetId string, number int) error {
	asset, _, err := modules.String2AssetId(assetId)
	if err != nil {
		return err
	}

	if number == 0 {
		newestUnitHash, newestIndex, _ := d.stablePropRep.GetNewestUnit(asset)
		log.Infof("Newest unit[%s] height:%d", newestUnitHash.String(), newestIndex.Index)
		number = int(newestIndex.Index)
	}

	header, err := d.stableUnitRep.GetHeaderByNumber(modules.NewChainIndex(asset, uint64(number)))
	if err != nil {
		return fmt.Errorf("Unit height:%d not exits", number)
	}

	txs, err1 := d.stableUnitRep.GetUnitTransactions(header.Hash())
	if err1 != nil {
		return fmt.Errorf("unit height:%d 's body not exist, maybe the txroot is incorrect.", number)
	}

	// check txroot
	root := core.DeriveSha(txs)
	if root != header.TxRoot() {
		return fmt.Errorf("unit height:%d 's txroot:%s is not equal to %s",
			number, header.TxRoot().String(), root.String())
	}
	parentHash := header.ParentHash()[0]
	parentNumber := header.NumberU64() - 1

	var incorrect_num string
	var incorrect_root bool
	for {
		header, err = d.stableUnitRep.GetHeaderByHash(parentHash)
		if err != nil {
			return fmt.Errorf("Unit :%s not exits", parentHash.String())
		}
		if header.NumberU64() != parentNumber {
			return fmt.Errorf("Number not correct,%d,%d", header.NumberU64(), parentNumber)
		}
		txs, err1 := d.stableUnitRep.GetUnitTransactions(header.Hash())
		if err1 != nil {
			return fmt.Errorf("unit height:%d 's body not exist, maybe the txroot is incorrect.",
				header.NumberU64())
		}
		// check txroot
		root := core.DeriveSha(txs)
		if root != header.TxRoot() {
			incorrect_root = true
			incorrect_num += fmt.Sprintf("%d,", header.NumberU64())
			log.Debugf("unit height[%d] 's txroot[%s] is not equal the correct[%s].", header.NumberU64(),
				header.TxRoot().String(), root.String())
		}

		if len(header.ParentHash()) > 0 {
			parentHash = header.ParentHash()[0]
			parentNumber = header.NumberU64() - 1
		} else {
			log.Infof("Check complete!%d", header.NumberU64())
			break
		}
		if header.NumberU64()%1000 == 0 {
			log.Infof("Check header correct:%d", header.NumberU64())
		}
	}

	if incorrect_root {
		return fmt.Errorf("unit height[%s] 's txroot is not equal the correct.", incorrect_num)
	}

	return nil
}
func (d *Dag) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	return d.unstableStateRep.GetBlacklistAddress()
}
func (d *Dag) RebuildAddrTxIndex() error {
	return d.stableUnitRep.RebuildAddrTxIndex()
}
func (d *Dag) GetJurorByAddrHash(hash common.Hash) (*modules.JurorDeposit, error) {
	return d.stableStateRep.GetJurorByAddrHash(hash)
}

func (d *Dag) MemdagInfos() (*modules.MemdagInfos, error) {
	ptn_memdag := d.Memdag
	memdag_infos := new(modules.MemdagInfos)
	memdag_infos.MemStatus = make(map[string]*modules.MemdagStatus)
	if ptn_status, err := ptn_memdag.Info(); err != nil {
		log.Errorf("get [%s] memdag status failed, error:[%s]", modules.PTNCOIN.String(), err.Error())
		return nil, err
	} else {
		memdag_infos.MemStatus[ptn_status.Token.String()] = ptn_status
	}
	for token, memdag := range d.PartitionMemDag {
		if token == modules.PTNCOIN {
			continue
		}
		if status, err := memdag.Info(); err == nil {
			memdag_infos.MemStatus[token.String()] = status
		} else {
			log.Errorf("get [%s] memdag status failed, error:[%s]", modules.PTNCOIN.String(), err.Error())
		}
	}
	return memdag_infos, nil
}

func (d *Dag) GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract {
	return d.stableStateRep.GetContractsWithJuryAddr(addr)
}
