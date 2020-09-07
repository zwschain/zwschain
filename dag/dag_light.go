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
 *  * @date 2018-2019
 *
 */

package dag

import (
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
)

// clear all utxo by address
func (dag *Dag) ClearUtxo() error {
	return dag.stableUtxoRep.ClearUtxo()
}
func (dag *Dag) ClearAddrUtxo(addr common.Address) error {
	return dag.stableUtxoRep.ClearAddrUtxo(addr)
}

// save all utxo of a view
func (dag *Dag) SaveUtxoView(view map[modules.OutPoint]*modules.Utxo) error {
	return dag.stableUtxoRep.SaveUtxoView(view)
}
