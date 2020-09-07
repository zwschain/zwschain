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

package core

import (
	"io"
	"strconv"

	"github.com/ethereum/go-ethereum/rlp"
)

// only for serialization(storage/p2p)
type ChainParametersTemp struct {
	ChainParametersBase

	ChainParametersExtraTemp
}

type ChainParametersExtraTemp struct {
	ChainParametersExtraTemp104alpha

	PledgeAllocateThreshold string
	PledgeRecordsThreshold  string
}

type ChainParametersExtraTemp104alpha struct {
	// TxCoinYearRate     string
	//DepositDailyReward string
	//DepositPeriod      string

	UccMemory string
	//UccMemorySwap string
	UccCpuShares string
	UccCpuQuota  string
	//UccCpuPeriod  string
	UccDisk       string
	UccDuringTime string

	TempUccMemory string
	//TempUccMemorySwap string
	TempUccCpuShares string
	TempUccCpuQuota  string

	ContractSystemVersion string
	ContractSignatureNum  string
	ContractElectionNum   string

	ContractTxTimeoutUnitFee  string
	ContractTxSizeUnitFee     string
	ContractTxInstallFeeLevel string
	ContractTxDeployFeeLevel  string
	ContractTxInvokeFeeLevel  string
	ContractTxStopFeeLevel    string
}

func (cp *ChainParameters) EncodeRLP(w io.Writer) error {
	cpt := cp.GetCPT()

	return rlp.Encode(w, cpt)
}

func (cp *ChainParametersExtra104alpha) GetCPT() *ChainParametersExtraTemp104alpha {
	return &ChainParametersExtraTemp104alpha{
		// TxCoinYearRate:     strconv.FormatFloat(float64(cp.TxCoinYearRate), 'f', -1, 64),
		//DepositDailyReward: strconv.FormatInt(int64(cp.PledgeDailyReward), 10),
		//DepositPeriod:      strconv.FormatInt(int64(cp.DepositPeriod), 10),

		UccMemory:     strconv.FormatInt(cp.UccMemory, 10),
		UccCpuShares:  strconv.FormatInt(cp.UccCpuShares, 10),
		UccCpuQuota:   strconv.FormatInt(cp.UccCpuQuota, 10),
		UccDisk:       strconv.FormatInt(cp.UccDisk, 10),
		UccDuringTime: strconv.FormatInt(cp.UccDuringTime, 10),

		TempUccMemory:    strconv.FormatInt(cp.TempUccMemory, 10),
		TempUccCpuShares: strconv.FormatInt(cp.TempUccCpuShares, 10),
		TempUccCpuQuota:  strconv.FormatInt(cp.TempUccCpuQuota, 10),

		ContractSystemVersion: cp.ContractSystemVersion,
		ContractSignatureNum:  strconv.FormatInt(int64(cp.ContractSignatureNum), 10),
		ContractElectionNum:   strconv.FormatInt(int64(cp.ContractElectionNum), 10),

		ContractTxTimeoutUnitFee:  strconv.FormatUint(cp.ContractTxTimeoutUnitFee, 10),
		ContractTxSizeUnitFee:     strconv.FormatUint(cp.ContractTxSizeUnitFee, 10),
		ContractTxInstallFeeLevel: strconv.FormatFloat(cp.ContractTxInstallFeeLevel, 'f', -1, 64),
		ContractTxDeployFeeLevel:  strconv.FormatFloat(cp.ContractTxDeployFeeLevel, 'f', -1, 64),
		ContractTxInvokeFeeLevel:  strconv.FormatFloat(cp.ContractTxInvokeFeeLevel, 'f', -1, 64),
		ContractTxStopFeeLevel:    strconv.FormatFloat(cp.ContractTxStopFeeLevel, 'f', -1, 64),
	}
}

func (cp *ChainParametersExtra) GetCPT() *ChainParametersExtraTemp {
	return &ChainParametersExtraTemp{
		ChainParametersExtraTemp104alpha: *cp.ChainParametersExtra104alpha.GetCPT(),

		PledgeAllocateThreshold: strconv.FormatInt(int64(cp.PledgeAllocateThreshold), 10),
		PledgeRecordsThreshold:  strconv.FormatInt(int64(cp.PledgeRecordsThreshold), 10),
	}
}

func (cp *ChainParameters) GetCPT() *ChainParametersTemp {
	return &ChainParametersTemp{
		ChainParametersBase: cp.ChainParametersBase,

		ChainParametersExtraTemp: *cp.ChainParametersExtra.GetCPT(),
	}
}

func (cpt *ChainParametersExtraTemp104alpha) GetCP(cp *ChainParametersExtra104alpha) error {
	// TxCoinYearRate, err := strconv.ParseFloat(cpt.TxCoinYearRate, 64)
	// if err != nil {
	// 	return err
	// }
	// cp.TxCoinYearRate = float64(TxCoinYearRate)

	//DepositDailyReward, err := strconv.ParseInt(cpt.DepositDailyReward, 10, 64)
	//if err != nil {
	//	return err
	//}
	//cp.PledgeDailyReward = uint64(DepositDailyReward)

	//DepositPeriod, err := strconv.ParseInt(cpt.DepositPeriod, 10, 64)
	//if err != nil {
	//	return err
	//}
	//cp.DepositPeriod = int(DepositPeriod)

	UccMemory, err := strconv.ParseInt(cpt.UccMemory, 10, 64)
	if err != nil {
		return err
	}
	cp.UccMemory = UccMemory

	UccCpuShares, err := strconv.ParseInt(cpt.UccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	cp.UccCpuShares = UccCpuShares

	UccCpuQuota, err := strconv.ParseInt(cpt.UccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	cp.UccCpuQuota = UccCpuQuota

	UccDisk, err := strconv.ParseInt(cpt.UccDisk, 10, 64)
	if err != nil {
		return err
	}
	cp.UccDisk = UccDisk

	UccDuringTime, err := strconv.ParseInt(cpt.UccDuringTime, 10, 64)
	if err != nil {
		return err
	}
	cp.UccDuringTime = UccDuringTime

	TempUccMemory, err := strconv.ParseInt(cpt.TempUccMemory, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccMemory = TempUccMemory

	TempUccCpuShares, err := strconv.ParseInt(cpt.TempUccCpuShares, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccCpuShares = TempUccCpuShares

	TempUccCpuQuota, err := strconv.ParseInt(cpt.TempUccCpuQuota, 10, 64)
	if err != nil {
		return err
	}
	cp.TempUccCpuQuota = TempUccCpuQuota

	ContractSignatureNum, err := strconv.ParseInt(cpt.ContractSignatureNum, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractSignatureNum = int(ContractSignatureNum)

	cp.ContractSystemVersion = cpt.ContractSystemVersion

	ContractElectionNum, err := strconv.ParseInt(cpt.ContractElectionNum, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractElectionNum = int(ContractElectionNum)

	ContractTxTimeoutUnitFee, err := strconv.ParseUint(cpt.ContractTxTimeoutUnitFee, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractTxTimeoutUnitFee = ContractTxTimeoutUnitFee

	ContractTxSizeUnitFee, err := strconv.ParseUint(cpt.ContractTxSizeUnitFee, 10, 64)
	if err != nil {
		return err
	}
	cp.ContractTxSizeUnitFee = ContractTxSizeUnitFee

	ContractTxInstallFeeLevel, err := strconv.ParseFloat(cpt.ContractTxInstallFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxInstallFeeLevel = ContractTxInstallFeeLevel

	ContractTxDeployFeeLevel, err := strconv.ParseFloat(cpt.ContractTxDeployFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxDeployFeeLevel = ContractTxDeployFeeLevel

	ContractTxInvokeFeeLevel, err := strconv.ParseFloat(cpt.ContractTxInvokeFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxInvokeFeeLevel = ContractTxInvokeFeeLevel

	ContractTxStopFeeLevel, err := strconv.ParseFloat(cpt.ContractTxStopFeeLevel, 64)
	if err != nil {
		return err
	}
	cp.ContractTxStopFeeLevel = ContractTxStopFeeLevel

	return nil
}

func (cpt *ChainParametersExtraTemp) GetCP(cp *ChainParametersExtra) error {
	err := cpt.ChainParametersExtraTemp104alpha.GetCP(&cp.ChainParametersExtra104alpha)
	if err != nil {
		return err
	}

	PledgeAllocateThreshold, err := strconv.ParseInt(cpt.PledgeAllocateThreshold, 10, 64)
	if err != nil {
		return err
	}
	cp.PledgeAllocateThreshold = int(PledgeAllocateThreshold)

	PledgeRecordsThreshold, err := strconv.ParseInt(cpt.PledgeRecordsThreshold, 10, 64)
	if err != nil {
		return err
	}
	cp.PledgeRecordsThreshold = int(PledgeRecordsThreshold)

	return nil
}

func (cpt *ChainParametersTemp) GetCP(cp *ChainParameters) error {
	cp.ChainParametersBase = cpt.ChainParametersBase

	err := cpt.ChainParametersExtraTemp.GetCP(&cp.ChainParametersExtra)
	if err != nil {
		return err
	}

	return nil
}

func (cp *ChainParameters) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}

	cpt := &ChainParametersTemp{}
	err = rlp.DecodeBytes(raw, cpt)
	if err != nil {
		return err
	}

	err = cpt.GetCP(cp)
	if err != nil {
		return err
	}

	return nil
}
