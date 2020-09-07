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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package dag

import (
	"encoding/json"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/event"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"go.dedis.ch/kyber/v3/sign/bls"
)

func (d *Dag) SubscribeToGroupSignEvent(ch chan<- modules.ToGroupSignEvent) event.Subscription {
	return d.Memdag.SubscribeToGroupSignEvent(ch)
}

func (d *Dag) IsActiveMediator(add common.Address) bool {
	return d.GetGlobalProp().IsActiveMediator(add)
}

func (d *Dag) IsPrecedingMediator(add common.Address) bool {
	return d.GetGlobalProp().IsPrecedingMediator(add)
}

func (dag *Dag) InitPropertyDB(genesis *core.Genesis, unit *modules.Unit) error {
	//  全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	gp := modules.InitGlobalProp(genesis)
	if err := dag.stablePropRep.StoreGlobalProp(gp); err != nil {
		return err
	}

	//  动态全局属性不是交易，不需要放在Unit中
	// @author Albert·Gou
	dgp := modules.InitDynGlobalProp()
	if err := dag.stablePropRep.StoreDynGlobalProp(dgp); err != nil {
		return err
	}
	//dag.stablePropRep.SetNewestUnit(unit.Header())

	//  初始化mediator调度器，并存在数据库
	// @author Albert·Gou
	ms := modules.InitMediatorSchl(gp, dgp)
	if err := dag.stablePropRep.StoreMediatorSchl(ms); err != nil {
		return err
	}
	dag.stablePropRep.UpdateMediatorSchedule()

	return nil
}

func (dag *Dag) InitStateDB(genesis *core.Genesis, head *modules.Header) error {
	version := &modules.StateVersion{
		Height:  head.GetNumber(),
		TxIndex: ^uint32(0),
	}

	// Create initial mediators
	list := make(map[string]bool, len(genesis.InitialMediatorCandidates))

	for _, imc := range genesis.InitialMediatorCandidates {
		// 1 存储 mediator info
		addr, jde, err := imc.Validate()
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		mi := modules.NewMediatorInfo()
		mi.MediatorInfoBase = imc.MediatorInfoBase

		err = dag.stableStateRep.StoreMediatorInfo(addr, mi)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		// 2 初始化mediator保证金设为0
		md := modules.NewMediatorDeposit()
		md.Status = modules.Agree
		md.Role = modules.Mediator
		md.ApplyEnterTime = time.Unix(head.Timestamp(), 0).UTC().Format(modules.Layout2)

		byte, err := json.Marshal(md)
		if err != nil {
			return err
		}

		ws := modules.NewWriteSet(storage.MediatorDepositKey(imc.AddStr), byte)
		err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		// 3 初始化 juror保证金为0
		juror := modules.JurorDeposit{}
		juror.Address = mi.AddStr
		juror.Role = modules.Jury
		juror.Balance = 0
		juror.EnterTime = md.ApplyEnterTime
		juror.JurorDepositExtra = jde

		jurorByte, err := json.Marshal(juror)
		if err != nil {
			log.Errorf(err.Error())
			return err
		}

		ws = modules.NewWriteSet(storage.JuryDepositKey(mi.AddStr), jurorByte)
		err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}

		// 加入 mediator和jury列表
		list[mi.AddStr] = true
	}

	// 存储 initMediatorCandidates/JuryCandidates
	imcB, err := json.Marshal(list)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	//Mediator
	ws := modules.NewWriteSet(modules.MediatorList, imcB)
	err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	//Jury
	ws.Key = modules.JuryList
	err = dag.stableStateRep.SaveContractState(syscontract.DepositContractAddress.Bytes(), ws, version)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	return nil
}

func (dag *Dag) IsSynced(toStrictly bool) bool {
	var now, nextSlotTime time.Time

	if toStrictly {
		nowFine := time.Now()
		now = time.Unix(nowFine.Add(500*time.Millisecond).Unix(), 0)
		nextSlotTime = dag.unstablePropRep.GetSlotTime(1)
	} else {
		// 防止误判，获取之后的第3个生产槽时间
		now = time.Now()
		nextSlotTime = dag.unstablePropRep.GetSlotTime(3)
	}

	return nextSlotTime.After(now)
}

// author Albert·Gou
func (d *Dag) ChainThreshold() int {
	return d.GetGlobalProp().ChainThreshold()
}

func (d *Dag) PrecedingThreshold() int {
	return d.GetGlobalProp().PrecedingThreshold()
}

func (d *Dag) UnitIrreversibleTime() time.Duration {
	gp := d.GetGlobalProp()
	cp := gp.ChainParameters
	it := uint(gp.ChainThreshold()+int(cp.MaintenanceSkipSlots)) * uint(cp.MediatorInterval)
	return time.Duration(it) * time.Second
}

func (d *Dag) IsIrreversibleUnit(hash common.Hash) (bool, error) {
	header, err := d.unstableUnitRep.GetHeaderByHash(hash)
	if err != nil {
		log.Debugf("UnitRep GetHeaderByHash error:%s", err.Error())
		return false, err // 不存在该unit
	}

	if header.NumberU64() > d.GetIrreversibleUnitNum(header.GetAssetId()) {
		return false, nil
	}

	return true, nil
}

func (d *Dag) GetIrreversibleUnitNum(id modules.AssetId) uint64 {
	_, idx, err := d.stablePropRep.GetNewestUnit(id)
	if err != nil {
		log.Debugf("stableUnitRep GetNewestUnit error:%s", err.Error())
		return 0
	}

	return idx.Index
}

func (d *Dag) VerifyUnitGroupSign(unitHash common.Hash, groupSign []byte) error {
	header, err := d.GetHeaderByHash(unitHash)
	if err != nil {
		log.Debugf("get header of unit(%v) err: %v", unitHash.TerminalString(), err.Error())
		return err
	}

	pubKey, err := header.GetGroupPubKey()
	if err != nil {
		log.Debugf("get pubKey of unit(%v) err: %v", unitHash.TerminalString(), err.Error())
		return err
	}

	err = bls.Verify(core.Suite, pubKey, unitHash[:], groupSign)
	if err != nil {
		log.Debugf("the group signature(%v)  of the unit(hash: %v , # %v )  have an error when verifying: %v",
			hexutil.Encode(groupSign), unitHash.TerminalString(), header.NumberU64(), err.Error())
		return err
	}
	return nil
}

// 判断该mediator是下一个产块mediator
func (dag *Dag) IsConsecutiveMediator(nextMediator common.Address) bool {
	dgp := dag.GetDynGlobalProp()

	if !dgp.IsShuffledSchedule && nextMediator.Equal(dgp.LastMediator) {
		return true
	}

	return false
}

// 计算最近128个生产slots的mediator参与度，不包括当前unit
// Calculate the percent of unit production slots that were missed in the
// past 128 units, not including the current unit.
func (dag *Dag) MediatorParticipationRate() uint32 {
	recentSlotsFilled := dag.GetDynGlobalProp().RecentSlotsFilled
	participationRate := core.PalletOne100Percent * int(recentSlotsFilled.PopCount()) / 128

	return uint32(participationRate)
}

// subscribe active mediators updated event
func (d *Dag) SubscribeActiveMediatorsUpdatedEvent(ch chan<- modules.ActiveMediatorsUpdatedEvent) event.Subscription {
	return d.unstableUnitProduceRep.SubscribeActiveMediatorsUpdatedEvent(ch)
}

func (d *Dag) SubscribeUnstableRepositoryUpdatedEvent(ch chan<- modules.UnstableRepositoryUpdatedEvent) event.Subscription {
	return d.unstableRepositoryUpdatedScope.Track(d.unstableRepositoryUpdatedFeed.Subscribe(ch))
}
