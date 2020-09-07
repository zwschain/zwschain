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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"

	"github.com/palletone/adaptor"
)

type PTNMain struct {
}

func (p *PTNMain) Init(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetStringArgs()
	if len(args) < 1 {
		return shim.Error("need 1 args (MapContractAddr)")
	}

	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(symbolsOwner, []byte(invokeAddr.String()))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}

	err = stub.PutState(symbolsContractMap, []byte(args[0]))
	if err != nil {
		return shim.Error("write symbolsContractMap failed: " + err.Error())
	}

	return shim.Success(nil)
}

func (p *PTNMain) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "setOwner":
		if len(args) < 1 {
			return shim.Error("need 1 args (PTNAddr)")
		}
		return p.SetOwner(args[0], stub)
	case "setETHContract":
		if len(args) < 1 {
			return shim.Error("need 1 args (MapContractAddr)")
		}
		return p.SetETHContract(args[0], stub)

	case "payoutPTNByETHAddr":
		if len(args) < 1 {
			return shim.Error("need 1 args (ETHAddr)")
		}
		return p.PayoutPTNByETHAddr(args[0], stub)
	case "payoutPTNByTxID":
		if len(args) < 1 {
			return shim.Error("need 1 args (ERC20TransferTxID)")
		}
		return p.PayoutPTNByTxID(args[0], stub)

	case "withdrawAmount":
		if len(args) < 2 {
			return shim.Error("need 2  args (PTNAddress,PTNAmount)")
		}
		withdrawAddr, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error("Invalid address string:" + args[0])
		}
		amount, err := decimal.NewFromString(args[1])
		if err != nil {
			return shim.Error("Invalid amount:" + args[1])
		}
		return p.WithdrawAmount(stub, withdrawAddr, amount)

	case "getPayout":
		if len(args) < 1 {
			return shim.Error("need 1 args (ERC20TransferTxID)")
		}
		return p.GetPayout(args[0], stub)

	case "get":
		if len(args) < 1 {
			return shim.Error("need 1 args (Key)")
		}
		return p.Get(stub, args[0])
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

//todo modify PTN ERC20 address
const PTN_ERC20Addr = "0xfe76be9cec465ed3219a9972c21655d57d21aec6"

//todo modify conforms 15
const Confirms = uint(15)

const symbolsOwner = "owner_"
const symbolsContractMap = "eth_map"

const symbolsPayout = "payout_"

func (p *PTNMain) SetOwner(ptnAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}
	err = stub.PutState(symbolsOwner, []byte(ptnAddr))
	if err != nil {
		return shim.Error("write symbolsOwner failed: " + err.Error())
	}
	return shim.Success([]byte("Success"))
}

func getOwner(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsOwner)
	if len(result) == 0 {
		return "", errors.New("Need set Owner")
	}

	return string(result), nil
}

func (p *PTNMain) SetETHContract(mapContractAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//only owner can set
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can set")
	}

	err = stub.PutState(symbolsContractMap, []byte(mapContractAddr))
	if err != nil {
		return shim.Error("write symbolsContractMap failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func getMapAddr(stub shim.ChaincodeStubInterface) (string, error) {
	result, _ := stub.GetState(symbolsContractMap)
	if len(result) == 0 {
		return "", errors.New("Need set MapContractAddr")
	}

	return string(result), nil
}

func (p *PTNMain) PayoutPTNByTxID(ethTxID string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	if "0x" == ethTxID[0:2] || "0X" == ethTxID[0:2] {
		ethTxID = ethTxID[2:]
	}
	result, _ := stub.GetState(symbolsPayout + ethTxID)
	if len(result) != 0 {
		log.Debugf("The tx has been payout")
		return shim.Error("The tx has been payout")
	}

	//get sender receiver amount
	txIDByte, err := hex.DecodeString(ethTxID)
	if err != nil {
		log.Debugf("txid invalid: %s", err.Error())
		return shim.Error(fmt.Sprintf("txid invalid: %s", err.Error()))
	}
	txResult, err := GetErc20Tx(txIDByte, stub)
	if err != nil {
		log.Debugf("GetErc20Tx failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//check tx status
	if !txResult.Tx.IsSuccess {
		log.Debugf("The tx is failed")
		return shim.Error("The tx is failed")
	}
	//check contract address, must be ptn erc20 contract address
	if strings.ToLower(txResult.Tx.TargetAddress) != PTN_ERC20Addr {
		log.Debugf("The tx is't ERC20 contract transfer of PTN")
		return shim.Error("The tx is't ERC20 contract transfer of PTN")
	}
	//check receiver, must be ptnmap contract address
	mapAddr, err := getMapAddr(stub)
	if err != nil {
		log.Debugf("getMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	if strings.ToLower(txResult.Tx.ToAddress) != mapAddr {
		log.Debugf("strings.ToLower(txResult.To): %s, mapAddr: %s ", strings.ToLower(txResult.Tx.ToAddress), mapAddr)
		return shim.Error("Not send token to the Map contract")
	}
	//check token amount
	bigIntAmount := txResult.Tx.Amount.Amount.Div(txResult.Tx.Amount.Amount, big.NewInt(1e10)) //Token's decimal is 18, PTN's decimal is 8
	amt := bigIntAmount.Uint64()
	if amt == 0 {
		log.Debugf("Amount is 0")
		return shim.Error("Amount is 0")
	}

	//check confirms
	curHeight, err := getHeight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}
	if curHeight-txResult.Tx.BlockHeight < Confirms {
		log.Debugf("Need more confirms")
		return shim.Error("Need more confirms")
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, txResult.Tx.FromAddress, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get addrPTN
	//ptnAddr := common.HexToAddress(ptnHex).String()
	//ptnAddr := "P" + base58.CheckEncode(common.FromHex(ptnHex), 0)
	if ptnAddr == "" {
		log.Debugf("Need transfer 1 PTNMap for bind address")
		return shim.Error("Need transfer 1 PTNMap for bind address")
	}
	//save payout history
	err = stub.PutState(symbolsPayout+ethTxID, []byte(ptnAddr+"-"+bigIntAmount.String()))
	if err != nil {
		log.Debugf("write symbolsPayout failed: %s", err.Error())
		return shim.Error("write symbolsPayout failed: " + err.Error())
	}

	//payout
	//asset := modules.NewPTNAsset()
	asset, _ := modules.StringToAsset("PTN")
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr, amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func (p *PTNMain) PayoutPTNByETHAddr(ethAddr string, stub shim.ChaincodeStubInterface) pb.Response {
	//
	mapAddr, err := getMapAddr(stub)
	if err != nil {
		log.Debugf("getMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get the mapping ptnAddr
	ptnAddr, err := getPTNMapAddr(mapAddr, ethAddr, stub)
	if err != nil {
		log.Debugf("getPTNMapAddr failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	//get addrPTN
	if ptnAddr == "" {
		log.Debugf("Need transfer 1 PTNMap for bind address")
		return shim.Error("Need transfer 1 PTNMap for bind address")
	}

	txResults, err := GetAddrHistory(ethAddr, mapAddr, stub)
	if err != nil {
		log.Debugf("GetAddrHistory failed: %s", err.Error())
		return shim.Error(err.Error())
	}

	curHeight, err := getHeight(stub)
	if curHeight == 0 || err != nil {
		return shim.Error("getHeight failed")
	}

	var amt uint64
	for _, txResult := range txResults.Txs {
		if !txResult.IsSuccess {
			continue
		}
		if 0 == txResult.Amount.Amount.Uint64() {
			continue
		}
		txIDHex := hex.EncodeToString(txResult.TxID)
		//check confirms
		if curHeight-txResult.BlockHeight < Confirms {
			log.Debugf("Need more confirms %s", txIDHex)
			continue
		}
		//
		result, _ := stub.GetState(symbolsPayout + txIDHex)
		if len(result) != 0 {
			log.Debugf("The tx %s has been payout", txIDHex)
			continue
		}
		log.Debugf("The tx %s need be payout", txIDHex)

		//check token amount
		bigIntAmout := txResult.Amount.Amount.Div(txResult.Amount.Amount, big.NewInt(1e10)) //Token's decimal is 18, PTN's decimal is 8
		amt += txResult.Amount.Amount.Uint64()

		//save payout history
		err = stub.PutState(symbolsPayout+txIDHex, []byte(ptnAddr+"-"+bigIntAmout.String()))
		if err != nil {
			log.Debugf("write symbolsPayout failed: %s", err.Error())
			return shim.Error("write symbolsPayout failed: " + err.Error())
		}

	}

	if amt == 0 {
		log.Debugf("Amount is 0")
		return shim.Error("Amount is 0")
	}

	//payout
	asset := modules.NewPTNAsset()
	amtToken := &modules.AmountAsset{Amount: amt, Asset: asset}
	err = stub.PayOutToken(ptnAddr, amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func GetErc20Tx(txID []byte, stub shim.ChaincodeStubInterface) (*adaptor.GetTransferTxOutput, error) {
	input := adaptor.GetTransferTxInput{TxID: txID}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetTransferTx", inputBytes)
	if err != nil {
		return nil, errors.New("GetTransferTx error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))

	//
	var output adaptor.GetTransferTxOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func GetAddrHistory(ethAddrFrom, mapAddrTo string, stub shim.ChaincodeStubInterface) (*adaptor.GetAddrTxHistoryOutput, error) {
	input := adaptor.GetAddrTxHistoryInput{FromAddress: ethAddrFrom, ToAddress: mapAddrTo, Asset: PTN_ERC20Addr,
		AddressLogicAndOr: true}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetAddrTxHistory", inputBytes)
	if err != nil {
		return nil, errors.New("GetAddrHistory error: " + err.Error())
	}
	log.Debugf("result : %s", string(result))
	//
	var output adaptor.GetAddrTxHistoryOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func getHeight(stub shim.ChaincodeStubInterface) (uint, error) {
	//
	input := adaptor.GetBlockInfoInput{Latest: true} //get best hight
	//
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	//adaptor.
	result, err := stub.OutChainCall("erc20", "GetBlockInfo", inputBytes)
	if err != nil {
		return 0, errors.New("GetBlockInfo error: " + err.Error())
	}
	//
	var output adaptor.GetBlockInfoOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return 0, err
	}

	if output.Block.BlockHeight == 0 {
		return 0, errors.New("{\"Error\":\"Failed to get eth height\"}")
	}

	return output.Block.BlockHeight, nil
}

func getPTNMapAddr(mapAddr, fromAddr string, stub shim.ChaincodeStubInterface) (string, error) {
	var input adaptor.GetPalletOneMappingAddressInput
	input.MappingDataSource = mapAddr
	input.ChainAddress = fromAddr

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	//
	result, err := stub.OutChainCall("erc20", "GetPalletOneMappingAddress", inputBytes)
	if err != nil {
		return "", errors.New("GetPalletOneMappingAddress failed: " + err.Error())
	}
	//
	var output adaptor.GetPalletOneMappingAddressOutput
	err = json.Unmarshal(result, &output)
	if err != nil {
		return "", err
	}
	if output.PalletOneAddress == "" {
		return "", errors.New("GetPalletOneMappingAddress result empty")
	}

	return output.PalletOneAddress, nil
}

func (p *PTNMain) GetPayout(ethTxID string, stub shim.ChaincodeStubInterface) pb.Response {
	if "0x" == ethTxID[0:2] || "0X" == ethTxID[0:2] {
		ethTxID = ethTxID[2:]
	}
	result, _ := stub.GetState(symbolsPayout + ethTxID)
	return shim.Success(result)
}

func (p *PTNMain) Get(stub shim.ChaincodeStubInterface, key string) pb.Response {
	result, _ := stub.GetState(key)
	return shim.Success(result)
}

func (p *PTNMain) WithdrawAmount(stub shim.ChaincodeStubInterface, withdrawAddr common.Address, amount decimal.Decimal) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}
	owner, err := getOwner(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if owner != invokeAddr.String() {
		return shim.Error("Only owner can withdraw")
	}

	//contractAddr
	amount = amount.Mul(decimal.New(100000000, 0))
	amtToken := &modules.AmountAsset{Amount: uint64(amount.IntPart()), Asset: modules.NewPTNAsset()}
	err = stub.PayOutToken(withdrawAddr.String(), amtToken, 0)
	if err != nil {
		log.Debugf("PayOutToken failed: %s", err.Error())
		return shim.Error("PayOutToken failed: " + err.Error())
	}

	return shim.Success([]byte("Success"))
}

func main() {
	err := shim.Start(new(PTNMain))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
