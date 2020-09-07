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

package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func shortId(id string) string {
	if len(id) < 8 {
		return id
	}
	return id[0:8]
}

/**
验证某个交易，tx具有以下规则：
Tx的第一条Msg必须是Payment
如果有ContractInvokeRequest，那么要么：
	1.	ContractInvoke不存在（这是一个Request）
	2.	ContractInvoke必然在Request的下面，不可能在Request的上面
	3.  不是Coinbase的情况下，创币PaymentMessage必须在Request下面，并由系统合约创建
	4.  如果是系统合约的请求和结果，必须重新运行合约，保证结果一致
To validate one transaction
如果isFullTx为false，意味着这个Tx还没有被陪审团处理完，所以结果部分的Payment不验证
*/
func (validate *Validate) validateTx(tx *modules.Transaction, isFullTx bool) (ValidationCode, []*modules.Addition) {
	if tx == nil {
		return TxValidationCode_VALID, nil
	}
	reqId := tx.RequestHash()
	msgs := tx.TxMessages()
	if len(msgs) == 0 {
		return TxValidationCode_INVALID_MSG, nil
	}
	isOrphanTx := false
	if msgs[0].App != modules.APP_PAYMENT { // 交易费
		return TxValidationCode_INVALID_MSG, nil
	}
	txFeePass, txFee := validate.validateTxFeeValid(tx)
	if !txFeePass {
		return TxValidationCode_INVALID_FEE, nil
	}
	// validate tx size
	if tx.Size().Float64() > float64(modules.TX_MAXSIZE) {
		log.Debugf("[%s]Tx size is to big.", shortId(reqId.String()))
		return TxValidationCode_NOT_COMPARE_SIZE, txFee
	}

	//合约的执行结果必须有Jury签名
	if validate.enableContractSignCheck && isFullTx && tx.IsContractTx() {
		if validate.enableContractRwSetCheck {
			if !validate.ContractTxCheck(tx) { //验证合约执行结果是够正常
				log.Debugf("[%s]ContractTxCheck fail", shortId(reqId.String()))
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
		}
		isResultMsg := false
		hasSignMsg := false
		for _, msg := range tx.TxMessages() {
			if msg.App.IsRequest() {
				isResultMsg = true
				continue
			}
			if isResultMsg && (msg.App == modules.APP_SIGNATURE || msg.App == modules.APP_PAYMENT) {
				hasSignMsg = true
			}
		}
		if !hasSignMsg {
			log.Warnf("[%s]tx is an user contract invoke, but don't have jury signature", shortId(reqId.String()))
			return TxValidationCode_INVALID_CONTRACT_SIGN, txFee
		}
	}
	hasRequestMsg := false
	requestMsgIndex := 9999
	isSysContractCall := false
	usedUtxo := make(map[string]bool) //Cached all used utxo in this tx
	for msgIdx, msg := range msgs {
		// check message type and payload
		if !validateMessageType(msg.App, msg.Payload) {
			return TxValidationCode_UNKNOWN_TX_TYPE, txFee
		}
		// validate every type payload
		switch msg.App {
		case modules.APP_PAYMENT:
			payment, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				return TxValidationCode_INVALID_PAYMMENTLOAD, txFee
			}
			//如果是合约执行结果中的Payment，只有是完整交易的情况下才检查解锁脚本
			if msgIdx > requestMsgIndex && !isFullTx {
				log.Debugf("[%s]tx is processing tx, don't need validate result payment", shortId(reqId.String()))
			} else {
				validateCode := validate.validatePaymentPayload(tx, msgIdx, payment, usedUtxo)
				if validateCode != TxValidationCode_VALID {
					if validateCode == TxValidationCode_ORPHAN {
						isOrphanTx = true
					} else {
						return validateCode, txFee
					}
				}
				//检查一个Tx是否包含了发币的Payment，如果有，那么检查是否是系统合约调用的结果
				if msgIdx != 0 && payment.IsCoinbase() && !isSysContractCall {
					log.Errorf("[%s]Invalid Coinbase message", shortId(reqId.String()))
					return TxValidationCode_INVALID_COINBASE, txFee
				}
			}
		case modules.APP_CONTRACT_TPL:
			payload, _ := msg.Payload.(*modules.ContractTplPayload)
			validateCode := validate.validateContractTplPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_DEPLOY:
			payload, _ := msg.Payload.(*modules.ContractDeployPayload)
			validateCode := validate.validateContractState(payload.ContractId, payload.ReadSet, payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_INVOKE:
			payload, _ := msg.Payload.(*modules.ContractInvokePayload)
			validateCode := validate.validateContractState(payload.ContractId, payload.ReadSet, payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_TPL_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			payload, _ := msg.Payload.(*modules.ContractInstallRequestPayload)
			if payload.TplName == "" || payload.Path == "" || payload.Version == "" {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			reqAddr, err := validate.dagquery.GetTxRequesterAddress(tx)
			if err != nil {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			if validate.enableDeveloperCheck {
				if !validate.statequery.IsContractDeveloper(reqAddr) {
					return TxValidationCode_NOT_TPL_DEVELOPER, txFee
				}
			}
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			// 参数临界值验证
			payload, _ := msg.Payload.(*modules.ContractDeployRequestPayload)
			if len(payload.TemplateId) == 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			validateCode := validate.validateContractDeploy(payload.TemplateId)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			if hasRequestMsg { //一个Tx只有一个Request
				return TxValidationCode_INVALID_MSG, txFee
			}
			hasRequestMsg = true
			requestMsgIndex = msgIdx
			payload, _ := msg.Payload.(*modules.ContractInvokeRequestPayload)
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			contractId := payload.ContractId
			if common.IsSystemContractId(contractId) {
				isSysContractCall = true
			}
		case modules.APP_CONTRACT_STOP_REQUEST:
			payload, _ := msg.Payload.(*modules.ContractStopRequestPayload)
			if len(payload.ContractId) == 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			// 验证ContractId有效性
			if len(payload.ContractId) <= 0 {
				return TxValidationCode_INVALID_CONTRACT, txFee
			}
			requestMsgIndex = msgIdx
		case modules.APP_CONTRACT_STOP:
			payload, _ := msg.Payload.(*modules.ContractStopPayload)
			validateCode := validate.validateContractState(payload.ContractId, payload.ReadSet, payload.WriteSet)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_SIGNATURE:
			// 签名验证
			payload, _ := msg.Payload.(*modules.SignaturePayload)
			validateCode := validate.validateContractSignature(payload.Signatures[:], tx, isFullTx)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_DATA:
			payload, _ := msg.Payload.(*modules.DataPayload)
			validateCode := validate.validateDataPayload(payload)
			if validateCode != TxValidationCode_VALID {
				return validateCode, txFee
			}
		case modules.APP_ACCOUNT_UPDATE:
			return validate.validateVoteMediatorTx(msg.Payload), txFee
		default:
			return TxValidationCode_UNKNOWN_TX_TYPE, txFee
		}
	}
	if isOrphanTx {
		return TxValidationCode_ORPHAN, txFee
	}
	return TxValidationCode_VALID, txFee
}

func (validate *Validate) validateVoteMediatorTx(payload interface{}) ValidationCode {
	accountUpdate, ok := payload.(*modules.AccountStateUpdatePayload)
	if !ok {
		log.Errorf("tx payload do not match type")
		return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
	}
	for _, writeSet := range accountUpdate.WriteSet {
		if writeSet.Key != constants.VOTED_MEDIATORS {
			continue
		}
		var mediators map[string]bool
		err := json.Unmarshal(writeSet.Value, &mediators)
		if err != nil {
			log.Errorf("writeSet value do not match key")
			return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
		}
		maxMediatorCount := int(validate.propquery.GetChainParameters().MaximumMediatorCount)
		mediatorCount := len(mediators)
		if mediatorCount > maxMediatorCount {
			log.Errorf("the total number(%v) of mediators voted exceeds the maximum limit: %v",
				mediatorCount, maxMediatorCount)
			return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
		}
		mp := validate.statequery.GetMediators()
		for mediatorStr, ok := range mediators {
			if !ok {
				log.Errorf("the value of map can only be true")
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}
			mediator, err := common.StringToAddress(mediatorStr)
			if err != nil {
				log.Errorf("invalid account address: %v", mediatorStr)
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}
			if !mp[mediator] {
				log.Errorf("%v is not mediator", mediatorStr)
				return TxValidationCode_UNSUPPORTED_TX_PAYLOAD
			}
		}
	}

	return TxValidationCode_VALID
}

//extSize :byte, extTime :s
func (validate *Validate) ValidateTxFeeEnough(tx *modules.Transaction, extSize float64, extTime float64) bool {
	if tx == nil {
		return false
	}
	if !validate.enableTxFeeCheck {
		return true
	}

	var onlyPayment = true
	var timeout uint32
	var opFee, sizeFee, timeFee, accountUpdateFee, appDataFee, allFee float64
	reqId := tx.RequestHash()
	txSize := tx.Size().Float64()

	if validate.propquery == nil || validate.utxoquery == nil {
		log.Warnf("[%s]ValidateTxFeeEnough, Cannot validate tx fee, your validate utxoquery or propquery not set", shortId(reqId.String()))
		return true //todo ?
	}
	fees, err := tx.GetTxFee(validate.utxoquery.GetUtxoEntry) //validate.dagquery.GetTxFee(tx)
	if err != nil {
		log.Errorf("[%s]validateTxFeeEnough, GetTxFee err:%s", shortId(reqId.String()), err.Error())
		return false //todo ?
	}

	cp := validate.propquery.GetChainParameters()
	timeUnitFee := float64(cp.ContractTxTimeoutUnitFee)
	sizeUnitFee := float64(cp.ContractTxSizeUnitFee)
	for _, msg := range tx.TxMessages() {
		switch msg.App {
		case modules.APP_CONTRACT_TPL_REQUEST:
			onlyPayment = false
			opFee = cp.ContractTxInstallFeeLevel
		case modules.APP_CONTRACT_DEPLOY_REQUEST:
			onlyPayment = false
			opFee = cp.ContractTxDeployFeeLevel
		case modules.APP_CONTRACT_INVOKE_REQUEST:
			onlyPayment = false
			opFee = cp.ContractTxInvokeFeeLevel
			timeout = msg.Payload.(*modules.ContractInvokeRequestPayload).Timeout
		case modules.APP_CONTRACT_STOP_REQUEST:
			onlyPayment = false
			opFee = cp.ContractTxStopFeeLevel
		case modules.APP_DATA:
			onlyPayment = false
			appDataFee = float64(cp.ChainParametersBase.TransferPtnPricePerKByte) * ((txSize + extSize) / 1024)
		case modules.APP_ACCOUNT_UPDATE:
			onlyPayment = false
			accountUpdateFee = float64(cp.ChainParametersBase.AccountUpdateFee)
		}
	}
	if onlyPayment {
		allFee = float64(cp.ChainParametersBase.TransferPtnBaseFee)
	} else {
		sizeFee = opFee * sizeUnitFee * (txSize + extSize)
		timeFee = opFee * timeUnitFee * (float64(timeout) + extTime)
		allFee = sizeFee + timeFee + accountUpdateFee + appDataFee
	}
	val := math.Max(float64(fees.Amount), allFee) == float64(fees.Amount)
	if !val {
		log.Errorf("[%s]validateTxFeeEnough invalid, fee amount[%f]-fees[%f] (%f + %f + %f + %f), "+
			"txSize[%f], timeout[%d], extSize[%f], extTime[%f]",
			shortId(reqId.String()), float64(fees.Amount), allFee, sizeFee, timeFee, accountUpdateFee, appDataFee,
			txSize, timeout, extSize, extTime)
	}

	log.Debugf("[%s]validateTxFeeEnough is %v, fee amount[%f]-fees[%f](%f + %f + %f + %f), "+
		"txSize[%f], timeout[%d], extSize[%f], extTime[%f]",
		shortId(reqId.String()), val, float64(fees.Amount), allFee, sizeFee, timeFee, accountUpdateFee, appDataFee,
		txSize, timeout, extSize, extTime)
	return val
}

//验证手续费是否合法，并返回手续费的分配情况
func (validate *Validate) validateTxFeeValid(tx *modules.Transaction) (bool, []*modules.Addition) {
	if tx == nil {
		log.Error("validateTxFeeValid, tx is nil")
		return false, nil
	}
	reqId := tx.RequestHash()
	if validate.utxoquery == nil {
		log.Warnf("[%s]validateTxFeeValid, Cannot validate tx fee, your validate utxoquery not set", shortId(reqId.String()))
		return true, nil
	}

	//check fee is or not enough
	assetId := dagconfig.DagConfig.GetGasToken()
	if validate.enableTxFeeCheck {
		if !validate.ValidateTxFeeEnough(tx, 0, 0) {
			log.Warnf("validateTxFeeValid, Tx[%s] fee is not enough", tx.Hash().String())
			return false, nil
		}
		feeAllocate, err := tx.GetTxFeeAllocate(validate.utxoquery.GetUtxoEntry,
			validate.tokenEngine.GetScriptSigners, common.Address{}, validate.statequery.GetJurorReward)
		if err != nil {
			log.Warnf("[%s]validateTxFeeValid, compute tx[%s] fee error:%s", shortId(reqId.String()), tx.Hash().String(), err.Error())
			return false, nil
		}
		//check fee type is ok

		for _, feeAsset := range feeAllocate {
			if feeAsset.Asset.String() != assetId.String() {
				log.Warnf("[%s]validateTxFeeValid, assetId is not equal, feeAsset:%s, cfg asset:%s", shortId(reqId.String()),
					feeAsset.Asset.String(), assetId.String())
				return false, feeAllocate
			}
		}

		return true, feeAllocate
	} else {
		feeAllocate, err := tx.GetTxFeeAllocateLegacyV1(validate.utxoquery.GetUtxoEntry,
			validate.tokenEngine.GetScriptSigners, common.Address{})
		if err != nil {
			log.Warnf("[%s]validateTxFeeValid, compute tx[%s] fee error:%s", shortId(reqId.String()), tx.Hash().String(), err.Error())
			return false, nil
		}
		//check fee type is ok
		for _, feeAsset := range feeAllocate {
			if feeAsset.Asset.String() != assetId.String() {
				log.Warnf("[%s]validateTxFeeValid, assetId is not equal, feeAsset:%s, cfg asset:%s", shortId(reqId.String()),
					feeAsset.Asset.String(), assetId.String())
				return false, feeAllocate
			}
		}
		return true, feeAllocate
	}

}

/**
检查message的app与payload是否一致
check messaage 'app' consistent with payload type
*/
func validateMessageType(app modules.MessageType, payload interface{}) bool {
	switch t := payload.(type) {
	case *modules.PaymentPayload:
		if app == modules.APP_PAYMENT {
			return true
		}
	case *modules.ContractTplPayload:
		if app == modules.APP_CONTRACT_TPL {
			return true
		}
	case *modules.ContractDeployPayload:
		if app == modules.APP_CONTRACT_DEPLOY {
			return true
		}
	case *modules.ContractInvokeRequestPayload:
		if app == modules.APP_CONTRACT_INVOKE_REQUEST {
			return true
		}
	case *modules.ContractInvokePayload:
		if app == modules.APP_CONTRACT_INVOKE {
			return true
		}
	case *modules.SignaturePayload:
		if app == modules.APP_SIGNATURE {
			return true
		}
	case *modules.DataPayload:
		if app == modules.APP_DATA {
			return true
		}
	case *modules.AccountStateUpdatePayload:
		if app == modules.APP_ACCOUNT_UPDATE {
			return true
		}
	case *modules.ContractDeployRequestPayload:
		if app == modules.APP_CONTRACT_DEPLOY_REQUEST {
			return true
		}
	case *modules.ContractInstallRequestPayload:
		if app == modules.APP_CONTRACT_TPL_REQUEST {
			return true
		}
	case *modules.ContractStopRequestPayload:
		if app == modules.APP_CONTRACT_STOP_REQUEST {
			return true
		}
	case *modules.ContractStopPayload:
		if app == modules.APP_CONTRACT_STOP {
			return true
		}

	default:
		log.Debug("The payload of message type is unexpected. ", "payload_type", t, "app type", app)
		return false
	}
	return false
}
func (validate *Validate) validateCoinbase(tx *modules.Transaction, ads []*modules.Addition) ValidationCode {
	contractId := syscontract.CoinbaseContractAddress.Bytes()
	msgs := tx.TxMessages()
	if msgs[0].App == modules.APP_PAYMENT { //到达一定高度，Account转UTXO

		//在Coinbase合约的StateDB中保存每个Mediator和Jury的奖励值，
		//key为奖励地址，Value为[]AmountAsset
		//读取之前的奖励统计值
		addrMap, err := validate.statequery.GetContractStatesByPrefix(contractId, constants.RewardAddressPrefix)
		if err != nil {
			return TxValidationCode_STATE_DATA_NOT_FOUND
		}
		rewards := map[common.Address][]modules.AmountAsset{}
		for key, v := range addrMap {
			addr := key[len(constants.RewardAddressPrefix):]
			incomeAddr, _ := common.StringToAddress(addr)
			var aa []modules.AmountAsset
			rlp.DecodeBytes(v.Value, &aa)
			if len(aa) > 0 {
				rewards[incomeAddr] = aa
			}
		}
		//附加最新的奖励
		for _, ad := range ads {

			reward, ok := rewards[ad.Addr]
			if !ok {
				reward = []modules.AmountAsset{}
			}
			reward = validate.addIncome(reward, ad.Amount, ad.Asset)
			rewards[ad.Addr] = reward
		}
		//Check payment output is correct
		payment := msgs[0].Payload.(*modules.PaymentPayload)
		if !validate.compareRewardAndOutput(rewards, payment.Outputs) {
			log.Errorf("Coinbase tx[%s] Output not match", tx.Hash().String())
			log.DebugDynamic(func() string {
				rjson, _ := json.Marshal(rewards)
				ojson, _ := json.Marshal(payment)
				return fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nPayment:%s", string(rjson), string(ojson))
			})
			// panic("Coinbase Output not match")
			return TxValidationCode_INVALID_COINBASE
		}
		//Check statedb should clear
		if len(addrMap) > 0 {
			clearStateInvoke := msgs[1].Payload.(*modules.ContractInvokePayload)
			if !bytes.Equal(clearStateInvoke.ContractId, contractId) {
				log.Errorf("Coinbase tx[%s] contract id not correct", tx.Hash().String())
				return TxValidationCode_INVALID_COINBASE
			}
			if !validate.compareRewardAndStateClear(rewards, clearStateInvoke.WriteSet) {
				rjson, _ := json.Marshal(rewards)
				ojson, _ := json.Marshal(clearStateInvoke)
				data := fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nInvoke result:%s", string(rjson), string(ojson))
				log.Errorf("Coinbase tx[%s] Clear statedb not match, detail data:%s",
					tx.Hash().String(), data)
				return TxValidationCode_INVALID_COINBASE
			}
		}
		return TxValidationCode_VALID
	}
	if msgs[0].App == modules.APP_CONTRACT_INVOKE { //Account模型记账
		//传入的ads,集合StateDB的历史，生成新的Reward记录
		rewards := map[common.Address][]modules.AmountAsset{}
		for _, v := range ads {
			key := constants.RewardAddressPrefix + v.Addr.String()
			data, version, err := validate.statequery.GetContractState(contractId, key)
			var income []modules.AmountAsset
			if err == nil { //之前有奖励
				rlp.DecodeBytes(data, &income)
			}
			//data = [] byte{}
			data, _ = json.Marshal(income)
			log.Debug(v.Addr.String() + ": Coinbase History reward:" + string(data) + " version:" + version.String())
			log.Debugf("Add reward %d %s to %s", v.Amount, v.Asset.String(), v.Addr.String())

			newValue := validate.addIncome(income, v.Amount, v.Asset)
			rewards[v.Addr] = newValue
		}
		//比对reward和writeset是否一致
		invoke := msgs[0].Payload.(*modules.ContractInvokePayload)
		if !bytes.Equal(invoke.ContractId, contractId) {
			log.Errorf("Coinbase tx[%s] contract id not correct", tx.Hash().String())
			return TxValidationCode_INVALID_COINBASE
		}
		if validate.compareRewardAndWriteset(rewards, invoke.WriteSet) {
			return TxValidationCode_VALID
		} else {
			rjson, _ := json.Marshal(rewards)
			ojson, _ := json.Marshal(invoke)
			var dbAa []modules.AmountAsset
			rlp.DecodeBytes(invoke.WriteSet[0].Value, &dbAa)
			aajson, _ := json.Marshal(dbAa)
			debugData := fmt.Sprintf("Data for help debug: \r\nRewards:%s \r\nInvoke result:%s, Writeset:%s",
				string(rjson), string(ojson), string(aajson))

			log.Errorf("Coinbase tx[%s] contract write set not correct, %s",
				tx.Hash().String(), debugData)
			return TxValidationCode_INVALID_COINBASE
		}
	}
	return TxValidationCode_VALID
}

func (validate *Validate) compareRewardAndOutput(rewards map[common.Address][]modules.AmountAsset, outputs []*modules.Output) bool {
	comparedCount := 0
	for addr, reward := range rewards {
		if validate.rewardExistInOutputs(addr, reward, outputs) {
			comparedCount++
		} else {
			return false
		}

	}
	return comparedCount == len(outputs)
	// if comparedCount != len(outputs) {
	// 	return false
	// }
	// return true
}
func (validate *Validate) rewardExistInOutputs(addr common.Address, aa []modules.AmountAsset, outputs []*modules.Output) bool {
	for _, out := range outputs {
		outAddr, _ := validate.tokenEngine.GetAddressFromScript(out.PkScript)
		if outAddr.Equal(addr) {

			for _, a := range aa {

				if a.Asset.Equal(out.Asset) && a.Amount != out.Value {
					return false
				}

			}
		}
	}
	return true
}
func (validate *Validate) compareRewardAndStateClear(rewards map[common.Address][]modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	comparedCount := 0
	empty, _ := rlp.EncodeToBytes([]modules.AmountAsset{})
	for addr := range rewards {
		addrKey := constants.RewardAddressPrefix + addr.String()
		for _, w := range writeset {
			// if !w.IsDelete {
			// 	return false
			// }
			if w.Key == addrKey && bytes.Equal(w.Value, empty) {
				comparedCount++
			}
		}

	}
	//return comparedCount == len(writeset)
	if comparedCount != len(rewards) { //所有的Reward的状态数据库被清空
		log.Warnf("write set comparedCount:%d clean count:%d", comparedCount, len(rewards))
		return false
	}
	return true
}
func (validate *Validate) compareRewardAndWriteset(rewards map[common.Address][]modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	comparedCount := 0
	for addr, reward := range rewards {

		if validate.rewardExist(addr, reward, writeset) {
			comparedCount++
		} else {

			return false
		}

	}
	return comparedCount == len(rewards)
	// if comparedCount != len(rewards) { //所有的Reward的状态数据库被清空
	// 	return false
	// }
	// return true
}
func (validate *Validate) rewardExist(addr common.Address, aa []modules.AmountAsset, writeset []modules.ContractWriteSet) bool {
	for _, w := range writeset {
		if w.Key == constants.RewardAddressPrefix+addr.String() {
			var dbAa []modules.AmountAsset
			err := rlp.DecodeBytes(w.Value, &dbAa)
			if err != nil {
				log.Error("Decode rlp data to []modules.AmountAsset error")
				return false
			}
			for _, a := range aa {
				for _, b := range dbAa {
					if a.Asset.Equal(b.Asset) && a.Amount != b.Amount {
						a1 := a
						b1 := b
						log.DebugDynamic(func() string {
							data, _ := json.Marshal(dbAa)
							return fmt.Sprintf("Coinbase rewardExist false, a[%d] b[%d], db writeset:%s", a1.Amount, b1.Amount, string(data))
						})
						return false
					}
				}
			}
		}
	}
	return true
}

func (validate *Validate) addIncome(income []modules.AmountAsset, newAmount uint64, asset *modules.Asset) []modules.AmountAsset {
	newValue := make([]modules.AmountAsset, 0, len(income))
	hasOldValue := false
	for _, aa := range income {
		if aa.Asset.Equal(asset) {
			aa.Amount += newAmount
			hasOldValue = true
		}
		newValue = append(newValue, aa)
	}
	if !hasOldValue {
		newValue = append(newValue, modules.AmountAsset{Amount: newAmount, Asset: asset})
	}
	return newValue
}
