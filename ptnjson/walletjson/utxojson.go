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

package walletjson

type UtxoJson struct {
	TxHash         string `json:"txid"`          // reference Utxo struct key field
	MessageIndex   uint32 `json:"message_index"` // message index in transaction
	OutIndex       uint32 `json:"out_index"`
	Amount         uint64 `json:"amount"`           // 数量
	Asset          string `json:"asset"`            // 资产类别
	PkScriptHex    string `json:"pk_script_hex"`    // 要执行的代码段
	PkScriptString string `json:"pk_script_string"` // 要执行的代码段
	LockTime       uint32 `json:"lock_time"`
}
