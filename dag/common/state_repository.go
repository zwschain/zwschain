/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"encoding/json"
	"errors"
	"github.com/palletone/go-palletone/common/util"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
)

type IStateRepository interface {
	GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error)
	GetContractStateByVersion(id []byte, field string, version *modules.StateVersion) ([]byte, error)

	SaveContract(contract *modules.Contract) error
	SaveContractState(id []byte, w *modules.ContractWriteSet, version *modules.StateVersion) error
	GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error)
	GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error)

	GetContract(id []byte) (*modules.Contract, error)
	GetAllContracts() ([]*modules.Contract, error)
	GetContractsByTpl(tplId []byte) ([]*modules.Contract, error)
	GetContractTpl(tplId []byte) (*modules.ContractTemplate, error)
	GetContractTplCode(tplId []byte) ([]byte, error)
	GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error)

	GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue, error)
	GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue, error)
	GetAccountBalance(address common.Address) uint64
	LookupAccount() map[common.Address]*modules.AccountInfo

	GetPledgeList() (*modules.PledgeList, error)
	GetMediatorVotedResults() (map[string]uint64, error)
	GetAccountVotedMediators(addr common.Address) map[string]bool
	GetVotingForMediator(addStr string) (map[string]uint64, error)

	GetMediator(add common.Address) *core.Mediator
	RetrieveMediator(address common.Address) (*core.Mediator, error)
	UpdateMediatorInfoExpand(med *core.Mediator) error
	GetMediators() map[common.Address]bool
	LookupMediatorInfo() []*modules.MediatorInfo
	IsMediator(address common.Address) bool
	RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error)
	StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error

	//GetCurrentChainIndex(assetId modules.AssetId) (*modules.ChainIndex, error)

	GetJuryCandidateList() (map[string]bool, error)
	IsJury(address common.Address) bool
	GetAllJuror() (map[string]*modules.JurorDeposit, error)
	//GetJurorByAddr(addr string) (*modules.JurorDeposit, error)
	GetJurorReward(jurorAdd common.Address) common.Address
	GetJurorByAddrHash(addrHash common.Hash) (*modules.JurorDeposit, error)
	GetContractDeveloperList() ([]common.Address, error)
	IsContractDeveloper(address common.Address) bool

	GetPartitionChains() ([]*modules.PartitionChain, error)
	GetMainChain() (*modules.MainChain, error)
	//获得一个合约的陪审团列表
	GetContractJury(contractId []byte) (*modules.ElectionNode, error)
	GetAllContractTpl() ([]*modules.ContractTemplate, error)
	GetDataVersion() (*modules.DataVersion, error)
	StoreDataVersion(dv *modules.DataVersion) error

	GetSysParamWithoutVote() (map[string]string, error)
	GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error)
	SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error
	GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error)

	SaveContractWithJuryAddr(addr common.Hash, contract *modules.Contract) error
	GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract
}

type StateRepository struct {
	statedb         storage.IStateDb
	dagdb           storage.IDagDb
	mapHash2Address map[common.Hash]common.Address //For Juror address hash
}

func NewStateRepository(statedb storage.IStateDb, dagdb storage.IDagDb) *StateRepository {
	return &StateRepository{statedb: statedb,
		mapHash2Address: make(map[common.Hash]common.Address),
		dagdb:           dagdb,
	}
}

func NewStateRepository4Db(db ptndb.Database) *StateRepository {
	statedb := storage.NewStateDb(db)
	dagdb := storage.NewDagDb(db)
	return NewStateRepository(statedb, dagdb)
}

//获取某个版本的值
//先根据StateVersion查到对应的交易
//然后交易的WriteSet找出来对应的key和value
func (rep *StateRepository) GetContractStateByVersion(id []byte,
	field string, version *modules.StateVersion) ([]byte, error) {
	unitHash, err := rep.dagdb.GetHashByNumber(version.Height)
	if err != nil {
		return nil, err
	}
	body, err := rep.dagdb.GetBody(unitHash)
	if err != nil {
		return nil, err
	}
	if len(body) <= int(version.TxIndex) {
		return nil, errors.New("tx body count less than version tx index")
	}
	txHash := body[version.TxIndex]
	tx, err := rep.dagdb.GetTransactionOnly(txHash)
	if err != nil {
		return nil, err
	}
	for _, msg := range tx.TxMessages() {
		if msg.App == modules.APP_CONTRACT_INVOKE {
			invoke := msg.Payload.(*modules.ContractInvokePayload)
			//if bytes.Equal(	invoke.ContractId,id){
			for _, write := range invoke.WriteSet {
				if write.Key == field {
					return write.Value, nil
				}
			}
			//}
		}
	}
	return nil, errors.New("WriteSet not found")
}

func (rep *StateRepository) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return rep.statedb.GetContractState(id, field)
}

func (rep *StateRepository) SaveSysConfigContract(key string, val []byte, ver *modules.StateVersion) error {
	return rep.statedb.SaveSysConfigContract(key, val, ver)
}

func (rep *StateRepository) GetSysParamWithoutVote() (map[string]string, error) {
	return rep.statedb.GetSysParamWithoutVote()
}

func (rep *StateRepository) GetSysParamsWithVotes() (*modules.SysTokenIDInfo, error) {
	return rep.statedb.GetSysParamsWithVotes()
}
func (rep *StateRepository) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	return rep.statedb.GetBlacklistAddress()
}
func (rep *StateRepository) GetContractStatesById(id []byte) (map[string]*modules.ContractStateValue, error) {
	return rep.statedb.GetContractStatesById(id)
}

func (rep *StateRepository) GetContractStatesByPrefix(id []byte,
	prefix string) (map[string]*modules.ContractStateValue, error) {
	return rep.statedb.GetContractStatesByPrefix(id, prefix)
}

func (rep *StateRepository) GetContract(id []byte) (*modules.Contract, error) {
	return rep.statedb.GetContract(id)
}

func (rep *StateRepository) SaveContract(contract *modules.Contract) error {
	return rep.statedb.SaveContract(contract)
}

func (rep *StateRepository) GetAllContracts() ([]*modules.Contract, error) {
	return rep.statedb.GetAllContracts()
}

func (rep *StateRepository) GetContractsByTpl(tplId []byte) ([]*modules.Contract, error) {
	cids, err := rep.statedb.GetContractIdsByTpl(tplId)
	if err != nil {
		return nil, err
	}
	result := make([]*modules.Contract, 0, len(cids))
	for _, cid := range cids {
		contract, err := rep.statedb.GetContract(cid)
		if err != nil {
			return nil, err
		}
		result = append(result, contract)
	}
	return result, nil
}

func (rep *StateRepository) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return rep.statedb.GetContractTpl(tplId)
}

func (rep *StateRepository) GetContractTplCode(tplId []byte) ([]byte, error) {
	return rep.statedb.GetContractTplCode(tplId)
}

func (rep *StateRepository) RetrieveMediator(address common.Address) (*core.Mediator, error) {
	return rep.statedb.RetrieveMediator(address)
}

func (rep *StateRepository) GetMediator(add common.Address) *core.Mediator {
	med, err := rep.statedb.RetrieveMediator(add)
	if err != nil {
		log.Debugf("Retrieve Mediator error: %v", err.Error())
		return nil
	}

	return med
}

func (rep *StateRepository) StoreMediator(med *core.Mediator) error {
	return rep.statedb.StoreMediator(med)
}

func (rep *StateRepository) UpdateMediatorInfoExpand(med *core.Mediator) error {
	return rep.statedb.UpdateMediatorInfoExpand(med)
}

func (rep *StateRepository) GetMediators() map[common.Address]bool {
	return rep.statedb.GetMediators()
}

func (rep *StateRepository) SaveContractState(contractId []byte, ws *modules.ContractWriteSet,
	version *modules.StateVersion) error {
	return rep.statedb.SaveContractState(contractId, ws, version)
}

func (rep *StateRepository) IsMediator(address common.Address) bool {
	return rep.statedb.IsMediator(address)
}

func (rep *StateRepository) GetAccountBalance(address common.Address) uint64 {
	return rep.statedb.GetAccountBalance(address)
}

func (rep *StateRepository) LookupAccount() map[common.Address]*modules.AccountInfo {
	return rep.statedb.LookupAccount()
}

func (rep *StateRepository) GetPledgeList() (*modules.PledgeList, error) {
	dd, _, err := rep.statedb.GetContractState(syscontract.DepositContractAddress.Bytes(), constants.PledgeListLastDate)
	if err != nil {
		return nil, err
	}
	date := string(dd)
	key := constants.PledgeList + date
	allM := &modules.PledgeList{}
	states, err := rep.statedb.GetContractStatesByPrefix(syscontract.DepositContractAddress.Bytes(),
		key)
	if err != nil {
		return nil, err
	}
	for _, v := range states {
		pledgeList := modules.PledgeList{}
		err = json.Unmarshal(v.Value, &pledgeList)
		if err != nil {
			log.Info("Unmarshal error: ", err.Error())
			return nil, err
		}
		allM.TotalAmount += pledgeList.TotalAmount
		allM.Members = append(allM.Members, pledgeList.Members...)
	}
	allM.Date = date
	return allM, nil
}

//获得新的用户的质押申请列表
func (rep *StateRepository) GetPledgeDepositApplyList() ([]*modules.AddressAmount, error) {
	states, err := rep.statedb.GetContractStatesByPrefix(syscontract.DepositContractAddress.Bytes(),
		string(constants.PLEDGE_DEPOSIT_PREFIX))
	if err != nil {
		return nil, err
	}
	result := []*modules.AddressAmount{}
	for _, v := range states {
		node := &modules.AddressAmount{}
		err = json.Unmarshal(v.Value, node)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}

func (rep *StateRepository) GetPledgeWithdrawApplyList() ([]*modules.AddressAmount, error) {
	states, err := rep.statedb.GetContractStatesByPrefix(syscontract.DepositContractAddress.Bytes(),
		string(constants.PLEDGE_WITHDRAW_PREFIX))
	if err != nil {
		return nil, err
	}
	result := []*modules.AddressAmount{}
	for _, v := range states {
		node := &modules.AddressAmount{}
		err = json.Unmarshal(v.Value, node)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}

//根据用户的新质押和提币申请，以及质押列表计算
func (rep *StateRepository) GetPledgeListWithNew() (*modules.PledgeList, error) {
	result := &modules.PledgeList{}

	pledgeList, _ := rep.GetPledgeList()
	if pledgeList != nil {
		result = pledgeList
	}

	newDepositList, _ := rep.GetPledgeDepositApplyList()
	for _, deposit := range newDepositList {
		result.Add(deposit.Address, deposit.Amount, 0)
	}

	//newWithdrawList, _ := rep.GetPledgeWithdrawApplyList()
	//for _, withdraw := range newWithdrawList {
	//	result.Reduce(withdraw.Address, withdraw.Amount)
	//}

	return result, nil
}

func (rep *StateRepository) GetMediatorVotedResults() (map[string]uint64, error) {
	mediatorVoteCount := make(map[string]uint64)

	mediators, err := rep.statedb.GetCandidateMediatorList()
	if err != nil {
		log.Debug("GetCandidateMediatorList error" + err.Error())
		return mediatorVoteCount, err
	}

	//先将所有mediator的投票数量设为0， 防止某个mediator未被任何账户投票
	for address := range mediators {
		mediatorVoteCount[address] = 0
	}

	pledgeList, err := rep.GetPledgeListWithNew()
	if err != nil {
		log.Warn("GetPledgeListWithNew error" + err.Error())
		return mediatorVoteCount, err
	}
	//log.DebugDynamic(func() string {
	//	data, _ := json.Marshal(pledgeList)
	//	return "GetPledgeListWithNew result:\r\n" + string(data)
	//})

	for _, account := range pledgeList.Members {
		// 遍历该账户投票的mediator
		addr, _ := common.StringToAddress(account.Address)
		votedMediators := rep.statedb.GetAccountVotedMediators(addr)
		for med := range votedMediators {
			// 判断账户投票的mediator是否仍为候选mediator
			if _, found := mediatorVoteCount[med]; !found {
				continue
			}

			// 累加投票数量
			mediatorVoteCount[med] += account.Amount
		}
	}

	return mediatorVoteCount, nil
}

func (rep *StateRepository) GetVotingForMediator(addStr string) (map[string]uint64, error) {
	votingMediatorCount := make(map[string]uint64)

	pledgeList, err := rep.GetPledgeListWithNew()
	if err != nil {
		log.Debug("GetPledgeListWithNew error" + err.Error())
		return votingMediatorCount, err
	}

	for _, account := range pledgeList.Members {
		// 遍历该账户投票的mediator
		addr, _ := common.StringToAddress(account.Address)
		votedMediators := rep.statedb.GetAccountVotedMediators(addr)
		for med := range votedMediators {
			// 判断该账户是否投票了指定的mediator
			if addStr == med {
				votingMediatorCount[account.Address] = account.Amount
				break
			}
		}
	}

	return votingMediatorCount, nil
}

func (rep *StateRepository) RetrieveMediatorInfo(address common.Address) (*modules.MediatorInfo, error) {
	return rep.statedb.RetrieveMediatorInfo(address)
}

func (rep *StateRepository) LookupMediatorInfo() []*modules.MediatorInfo {
	return rep.statedb.LookupMediatorInfo()
}

func (rep *StateRepository) StoreMediatorInfo(add common.Address, mi *modules.MediatorInfo) error {
	return rep.statedb.StoreMediatorInfo(add, mi)
}

func (rep *StateRepository) GetContractDeploy(tempId, contractId []byte, name string) (*modules.ContractDeployPayload, error) {
	return rep.statedb.GetContractDeploy(tempId[:])
}

func (rep *StateRepository) GetJuryCandidateList() (map[string]bool, error) {
	return rep.statedb.GetJuryCandidateList()
}

func (rep *StateRepository) GetJurorReward(jurorAdd common.Address) common.Address {
	jd, err := rep.statedb.GetJurorByAddr(jurorAdd.Str())
	if err != nil {
		log.Infof("Juror[%s] don't have reward address, use default account address.error:%s", jurorAdd.String(), err.Error())
		return jurorAdd
	}

	return jd.GetRewardAdd()
}

func (rep *StateRepository) GetJurorByAddrHash(hash common.Hash) (*modules.JurorDeposit, error) {
	if addr, exist := rep.mapHash2Address[hash]; exist {
		log.Debugf("GetJurorByAddrHash(hash:%s) in cache map,addr:%s",
			hash.String(), addr.String())
		return rep.statedb.GetJurorByAddr(addr.String())
	}
	//Not exist
	jurors, err := rep.GetAllJuror()
	if err != nil {
		log.Warn("GetAllJuror return error:%s", err.Error())
		return nil, err
	}
	//data, _ := json.Marshal(jurors)
	//log.Debugf("Jurors:%s", string(data))
	var result *modules.JurorDeposit
	for _, j := range jurors {
		jaddr, _ := common.StringToAddress(j.Address)
		jhash := util.RlpHash(jaddr)
		rep.mapHash2Address[jhash] = jaddr
		if jhash == hash {
			result = j
		}
	}
	if result == nil {
		return nil, errors.New("juror not found by hash:" + hash.String())
	}
	return result, nil
}
func (rep *StateRepository) GetAllJuror() (map[string]*modules.JurorDeposit, error) {
	return rep.statedb.GetAllJuror()
}

func (rep *StateRepository) IsJury(address common.Address) bool {
	return rep.statedb.IsInJuryCandidateList(address)
}

func (rep *StateRepository) GetContractDeveloperList() ([]common.Address, error) {
	return rep.statedb.GetContractDeveloperList()
}

func (rep *StateRepository) IsContractDeveloper(address common.Address) bool {
	return rep.statedb.IsInContractDeveloperList(address)
}

func (rep *StateRepository) GetPartitionChains() ([]*modules.PartitionChain, error) {
	return rep.statedb.GetPartitionChains()
}

func (rep *StateRepository) GetMainChain() (*modules.MainChain, error) {
	return rep.statedb.GetMainChain()
}

func (rep *StateRepository) GetAllAccountStates(address common.Address) (map[string]*modules.ContractStateValue,
	error) {
	return rep.statedb.GetAllAccountStates(address)
}

func (rep *StateRepository) GetAccountState(address common.Address, statekey string) (*modules.ContractStateValue,
	error) {
	return rep.statedb.GetAccountState(address, statekey)
}

//获得一个合约的陪审团列表
func (rep *StateRepository) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	return rep.statedb.GetContractJury(contractId)
}
func (rep *StateRepository) GetAllContractTpl() ([]*modules.ContractTemplate, error) {
	return rep.statedb.GetAllContractTpl()
}

func (rep *StateRepository) GetAccountVotedMediators(addr common.Address) map[string]bool {
	return rep.statedb.GetAccountVotedMediators(addr)
}

func (rep *StateRepository) GetDataVersion() (*modules.DataVersion, error) {
	return rep.statedb.GetDataVersion()
}
func (rep *StateRepository) StoreDataVersion(dv *modules.DataVersion) error {
	return rep.statedb.SaveDataVersion(dv)
}
func (rep *StateRepository) SaveContractWithJuryAddr(addr common.Hash, contract *modules.Contract) error {
	return rep.statedb.SaveContractWithJuryAddr(addr, contract)
}

func (rep *StateRepository) GetContractsWithJuryAddr(addr common.Hash) []*modules.Contract {
	return rep.statedb.GetContractsWithJuryAddr(addr)
}
