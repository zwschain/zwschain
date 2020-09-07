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
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
)

func (gp *GlobalProperty) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}

	gpt := &GlobalPropertyTemp{}
	err = rlp.DecodeBytes(raw, gpt)
	if err != nil {
		return err
	}

	err = gpt.getGP(gp)
	if err != nil {
		return err
	}

	return nil
}

func (gp *GlobalProperty) EncodeRLP(w io.Writer) error {
	temp := gp.getGPT()
	return rlp.Encode(w, temp)
}

// only for serialization(storage/p2p)
type GlobalPropertyTemp struct {
	GlobalPropBaseTemp
	GlobalPropExtraTemp
}

type GlobalPropExtraTemp struct {
	//ActiveJuries       []common.Address
	ActiveMediators    []common.Address
	PrecedingMediators []common.Address
}

type GlobalPropBaseTemp struct {
	ImmutableParameters core.ImmutableChainParameters
	ChainParametersTemp core.ChainParametersTemp
}

func (gp *GlobalProperty) getGPT() *GlobalPropertyTemp {
	gpbt := GlobalPropBaseTemp{
		ImmutableParameters: gp.ImmutableParameters,
		ChainParametersTemp: *gp.ChainParameters.GetCPT(),
	}

	gpet := GlobalPropExtraTemp{
		//ActiveJuries:       make([]common.Address, 0, len(gp.ActiveJuries)),
		ActiveMediators:    make([]common.Address, 0, len(gp.ActiveMediators)),
		PrecedingMediators: make([]common.Address, 0, len(gp.PrecedingMediators)),
	}

	gpt := &GlobalPropertyTemp{
		GlobalPropBaseTemp:  gpbt,
		GlobalPropExtraTemp: gpet,
	}

	//for juryAdd := range gp.ActiveJuries {
	//	gpt.ActiveJuries = append(gpt.ActiveJuries, juryAdd)
	//}

	for medAdd := range gp.ActiveMediators {
		gpt.ActiveMediators = append(gpt.ActiveMediators, medAdd)
	}

	for medAdd := range gp.PrecedingMediators {
		gpt.PrecedingMediators = append(gpt.PrecedingMediators, medAdd)
	}

	return gpt
}

func (gpt *GlobalPropertyTemp) getGP(gp *GlobalProperty) error {
	//gp.ActiveJuries = make(map[common.Address]bool)
	gp.ActiveMediators = make(map[common.Address]bool)
	gp.PrecedingMediators = make(map[common.Address]bool)

	//for _, addStr := range gpt.ActiveJuries {
	//	gp.ActiveJuries[addStr] = true
	//}

	for _, addStr := range gpt.ActiveMediators {
		gp.ActiveMediators[addStr] = true
	}

	for _, addStr := range gpt.PrecedingMediators {
		gp.PrecedingMediators[addStr] = true
	}

	gp.ImmutableParameters = gpt.ImmutableParameters

	return gpt.ChainParametersTemp.GetCP(&gp.ChainParameters)
}
