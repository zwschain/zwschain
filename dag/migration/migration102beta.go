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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018-2019
 *
 */
package migration

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/storage"
)

type Migration101_102 struct {
	dagdb   ptndb.Database
	idxdb   ptndb.Database
	utxodb  ptndb.Database
	statedb ptndb.Database
	propdb  ptndb.Database
}

func (m *Migration101_102) FromVersion() string {
	return "1.0.1-beta"
}

func (m *Migration101_102) ToVersion() string {
	return "1.0.2-beta"
}

func (m *Migration101_102) ExecuteUpgrade() error {
	//转换GLOBALPROPERTY结构体
	if err := m.upgradeGP(); err != nil {
		return err
	}

	return nil
}

func (m *Migration101_102) upgradeGP() error {
	oldGp := &GlobalProperty101{}
	err := storage.RetrieveFromRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, oldGp)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	newData := &GlobalProperty102delta{}
	newData.ActiveJuries = oldGp.ActiveJuries
	newData.ActiveMediators = oldGp.ActiveMediators
	newData.PrecedingMediators = oldGp.PrecedingMediators
	newData.ChainParameters = oldGp.ChainParameters

	newData.ImmutableParameters.MinMaintSkipSlots = 2
	newData.ImmutableParameters.MinimumMediatorCount = oldGp.ImmutableParameters.MinimumMediatorCount
	newData.ImmutableParameters.MinMediatorInterval = oldGp.ImmutableParameters.MinMediatorInterval
	newData.ImmutableParameters.UccPrivileged = oldGp.ImmutableParameters.UccPrivileged
	newData.ImmutableParameters.UccCapDrop = oldGp.ImmutableParameters.UccCapDrop
	newData.ImmutableParameters.UccNetworkMode = oldGp.ImmutableParameters.UccNetworkMode
	newData.ImmutableParameters.UccOOMKillDisable = oldGp.ImmutableParameters.UccOOMKillDisable

	err = storage.StoreToRlpBytes(m.propdb, constants.GLOBALPROPERTY_KEY, newData)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

type GlobalProperty101 struct {
	GlobalPropBase101

	ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type GlobalPropBase101 struct {
	ImmutableParameters ImmutableChainParameters101 // 不可改变的区块链网络参数
	ChainParameters     ChainParameters102delta     // 区块链网络参数
}

type ImmutableChainParameters101 struct {
	MinimumMediatorCount uint8    `json:"min_mediator_count"`    // 最小活跃mediator数量
	MinMediatorInterval  uint8    `json:"min_mediator_interval"` // 最小的生产槽间隔时间
	UccPrivileged        bool     `json:"ucc_privileged"`        // 防止容器以root权限运行
	UccCapDrop           []string `json:"ucc_cap_drop"`          // 确保容器以最小权限运行
	UccNetworkMode       string   `json:"ucc_network_mode"`      // 容器运行网络模式
	UccOOMKillDisable    bool     `json:"ucc_oom_kill_disable"`  // 是否内存使用量超过上限时系统杀死进程
}
