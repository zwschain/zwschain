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

package modules

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"io"
)

type ContractDeployPayloadV1 struct {
	TemplateId []byte             `json:"template_id"`    // contract template id
	ContractId []byte             `json:"contract_id"`    // contract id
	Name       string             `json:"name"`           // the name for contract
	Args       [][]byte           `json:"args"`           // contract arguments list
	EleList    []ElectionInf      `json:"election_list"`  // contract jurors list
	ReadSet    []ContractReadSet  `json:"read_set"`       // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`      // the set data of write, and value could be any type
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}
type ContractDeployPayloadV2 struct {
	TemplateId []byte             `json:"template_id"`   // delete--
	ContractId []byte             `json:"contract_id"`   // contract id
	Name       string             `json:"name"`          // the name for contract
	Args       [][]byte           `json:"args"`          // delete--
	EleNode    ElectionNode       `json:"election_node"` // contract jurors node info
	ReadSet    []ContractReadSet  `json:"read_set"`      // the set data of read, and value could be any type
	WriteSet   []ContractWriteSet `json:"write_set"`     // the set data of write, and value could be any type
	DuringTime uint64             `json:"during_time"`
	ErrMsg     ContractError      `json:"contract_error"` // contract error message
}

func (input *ContractDeployPayload) EncodeRLP(w io.Writer) error {

	if common.IsSystemContractId(input.ContractId) { //系统合约
		log.Debugf("System contract[%x] deploy payload rlp", input.ContractId)
		temp := &ContractDeployPayloadV1{}
		temp.TemplateId = input.TemplateId
		temp.ContractId = input.ContractId
		temp.Name = input.Name
		temp.Args = input.Args
		temp.EleList = input.EleNode.EleList
		temp.ReadSet = input.ReadSet
		temp.WriteSet = input.WriteSet
		temp.ErrMsg = input.ErrMsg
		return rlp.Encode(w, temp)
	}
	temp := &ContractDeployPayloadV2{}
	temp.TemplateId = input.TemplateId
	temp.ContractId = input.ContractId
	temp.Name = input.Name
	temp.Args = input.Args
	temp.EleNode = input.EleNode
	temp.ReadSet = input.ReadSet
	temp.WriteSet = input.WriteSet
	temp.DuringTime = input.DuringTime
	temp.ErrMsg = input.ErrMsg
	return rlp.Encode(w, temp)
}
