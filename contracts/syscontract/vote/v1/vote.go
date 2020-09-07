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

package v1

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	dm "github.com/palletone/go-palletone/dag/modules"
)

type GlobalTokenInfo struct {
	Symbol      string
	TokenType   uint8 //1:prc20 2:prc721 3:vote 4:SysVote
	Status      uint8
	CreateAddr  string
	TotalSupply uint64
	SupplyAddr  string
	AssetID     dm.AssetId
}

const symbolsKey = "symbol_"
const jsonResp1 = "{\"Error\":\"Results format invalid, Error!!!\"}"
const jsonResp2 = "{\"Error\":\"Token not exist\"}"

//Vote
type Vote struct {
}

//one topic
type voteTopic struct {
	TopicTitle    string
	SelectOptions []string
	SelectMax     uint64
}
type voteResult struct {
	SelectOption string
	Num          uint64
}

//topic support result
type topicSupports struct {
	TopicTitle  string
	VoteResults []voteResult
	SelectMax   uint64
	//SelectOptionsNum  uint64
}

//vote token information
type tokenInfo struct {
	Name        string
	Symbol      string
	CreateAddr  string
	VoteType    byte
	TotalSupply uint64
	VoteEndTime time.Time
	VoteContent []byte
	AssetID     dm.AssetId
}

//one user's support
type supportRequest struct {
	TopicIndex   uint64
	SelectIndexs []uint64
}

//Init
func (v *Vote) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke
func (v *Vote) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, args := stub.GetFunctionAndParameters()

	switch f {
	case "createToken":
		if len(args) < 5 {
			return shim.Error("need 5 args (Name,VoteType,TotalSupply,VoteEndTime,VoteContentJson)")
		}
		totalSupply, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to convert total supply\"}"
			return shim.Error(jsonResp)
		}
		if totalSupply == 0 {
			jsonResp := "{\"Error\":\"Can't be zero\"}"
			return shim.Error(jsonResp)
		}
		return v.CreateToken(stub, args[0], args[1], totalSupply, args[3], args[4])
	case "support":
		if len(args) < 1 {
			return shim.Error("need 1 args (SupportRequestJson)")
		}
		return v.Support(stub, args[0])
	case "getVoteResult":
		if len(args) < 1 {
			return shim.Error("need 1 args (AssetID)")
		}
		tkIDInfo, err0 := v.GetVoteResult(stub, args[0])
		if err0 != nil {
			return shim.Error(err0.Error())
		}
		tkJSON, err := json.Marshal(tkIDInfo)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(tkJSON)
	case "getVoteInfo":
		if len(args) < 1 {
			return shim.Error("need 1 args (AssetID)")
		}
		voteInfo, err0 := v.GetVoteInfo(stub, args[0])
		if err0 != nil {
			return shim.Error(err0.Error())
		}
		vtJSON, err := json.Marshal(voteInfo)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(vtJSON)
	default:
		jsonResp := "{\"Error\":\"Unknown function " + f + "\"}"
		return shim.Error(jsonResp)
	}
}

func setGlobal(stub shim.ChaincodeStubInterface, tkInfo *tokenInfo) error {
	gTkInfo := GlobalTokenInfo{Symbol: tkInfo.Symbol, TokenType: 3, Status: 0, CreateAddr: tkInfo.CreateAddr,
		TotalSupply: tkInfo.TotalSupply, SupplyAddr: "", AssetID: tkInfo.AssetID}
	val, err := json.Marshal(gTkInfo)
	if err != nil {
		return err
	}
	err = stub.PutGlobalState(dm.GlobalPrefix+gTkInfo.Symbol, val)
	return err
}

func setSymbols(stub shim.ChaincodeStubInterface, tkInfo *tokenInfo) error {
	val, err := json.Marshal(tkInfo)
	if err != nil {
		return err
	}
	err = stub.PutState(symbolsKey+tkInfo.AssetID.String(), val)
	return err
}
func getSymbols(stub shim.ChaincodeStubInterface, assetID string) *tokenInfo {
	//
	tkInfo := tokenInfo{}
	tkInfoBytes, _ := stub.GetState(symbolsKey + assetID)
	if len(tkInfoBytes) == 0 {
		return nil
	}
	//
	err := json.Unmarshal(tkInfoBytes, &tkInfo)
	if err != nil {
		return nil
	}

	return &tkInfo
}

func (v *Vote) CreateToken(stub shim.ChaincodeStubInterface, name string, voteType string, totalSupply uint64,
	voteEndTime string, voteContentJSON string) pb.Response {
	//==== convert params to token information
	var vt dm.VoteToken
	//name symbol
	if len(name) > 1024 {
		jsonResp := "{\"Error\":\"Name length should not be greater than 1024\"}"
		return shim.Error(jsonResp)
	}
	vt.Name = name
	vt.Symbol = "VOTE"

	//vote type
	if voteType == "0" {
		vt.VoteType = byte(0)
	} else if voteType == "1" {
		vt.VoteType = byte(1)
	} else if voteType == "2" {
		vt.VoteType = byte(2)
	} else {
		jsonResp := "{\"Error\":\"Only string, 0 or 1 or 2\"}"
		return shim.Error(jsonResp)
	}
	//total supply
	vt.TotalSupply = totalSupply
	//VoteEndTime
	endTime, err := time.Parse("2006-01-02 15:04:05", voteEndTime)
	if err != nil {
		jsonResp := "{\"Error\":\"No vote end time\"}"
		return shim.Error(jsonResp)
	}
	vt.VoteEndTime = endTime
	//VoteContent
	var voteTopics []voteTopic
	err = json.Unmarshal([]byte(voteContentJSON), &voteTopics)
	if err != nil {
		jsonResp := "{\"Error\":\"VoteContent format invalid\"}"
		return shim.Error(jsonResp)
	}
	//init support
	supports := make([]topicSupports, 0, len(voteTopics))
	for _, oneTopic := range voteTopics {
		var oneSupport topicSupports
		oneSupport.TopicTitle = oneTopic.TopicTitle
		for _, oneOption := range oneTopic.SelectOptions {
			var oneResult voteResult
			oneResult.SelectOption = oneOption
			oneSupport.VoteResults = append(oneSupport.VoteResults, oneResult)
		}
		//oneResult.SelectOptionsNum = uint64(len(oneRequest.SelectOptions))
		if oneTopic.SelectMax > uint64(len(oneTopic.SelectOptions)) {
			jsonResp := "{\"Error\":\"VoteContent 's SelectMax invalid\"}"
			return shim.Error(jsonResp)
		}
		oneSupport.SelectMax = oneTopic.SelectMax
		supports = append(supports, oneSupport)
	}
	voteContentJson, err := json.Marshal(supports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return shim.Error(jsonResp)
	}
	vt.VoteContent = voteContentJson

	txid := stub.GetTxID()
	assetID, _ := dm.NewAssetId(vt.Symbol, dm.AssetType_VoteToken,
		0, common.Hex2Bytes(txid[2:]), dm.UniqueIdType_Null)
	assetIDStr := assetID.String()
	//check name is only or not
	tkInfo := getSymbols(stub, assetIDStr)
	if tkInfo != nil {
		jsonResp := "{\"Error\":\"Repeat AssetID\"}"
		return shim.Error(jsonResp)
	}

	//convert to json
	createJson, err := json.Marshal(vt)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate token Json\"}"
		return shim.Error(jsonResp)
	}
	//get creator
	createAddr, err := stub.GetInvokeAddress()
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get invoke address\"}"
		return shim.Error(jsonResp)
	}

	//last put state
	info := tokenInfo{vt.Name, vt.Symbol, createAddr.String(), vt.VoteType, totalSupply,
		endTime, voteContentJson, assetID}

	err = setSymbols(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}

	//set token define
	err = stub.DefineToken(byte(dm.AssetType_VoteToken), createJson, createAddr.String())
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to call stub.DefineToken\"}"
		return shim.Error(jsonResp)
	}

	//add global state
	err = setGlobal(stub, &info)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to add global state\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(createJson)
}

func (v *Vote) Support(stub shim.ChaincodeStubInterface, supportRequestJSON string) pb.Response {
	//check token
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		jsonResp := "{\"Error\":\"GetInvokeTokens failed\"}"
		return shim.Error(jsonResp)
	}
	voteNum := uint64(0)
	assetIDStr := ""
	for i := 0; i < len(invokeTokens); i++ {
		if invokeTokens[i].Asset.AssetId == dm.PTNCOIN {
			continue
		} else if invokeTokens[i].Address == "P1111111111111111111114oLvT2" {
			if assetIDStr == "" {
				assetIDStr = invokeTokens[i].Asset.String()
				voteNum += invokeTokens[i].Amount
			} else if invokeTokens[i].Asset.AssetId.String() == assetIDStr {
				voteNum += invokeTokens[i].Amount
			}
		}
	}
	if voteNum == 0 || assetIDStr == "" { //no vote token
		jsonResp := "{\"Error\":\"Vote token empty\"}"
		return shim.Error(jsonResp)
	}

	//check name is exist or not
	tkInfo := getSymbols(stub, assetIDStr)
	if tkInfo == nil {
		return shim.Error(jsonResp2)
	}

	//parse support requests
	var requests []supportRequest
	err = json.Unmarshal([]byte(supportRequestJSON), &requests)
	if err != nil {
		jsonResp := "{\"Error\":\"SupportRequestJson format invalid\"}"
		return shim.Error(jsonResp)
	}
	//get token information
	var topicSupports []topicSupports
	err = json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return shim.Error(jsonResp1)
	}

	//if voteNum < uint64(len(supportRequests)) { //vote token more than request
	//	jsonResp := "{\"Error\":\"Vote token more than support request\"}"
	//	return shim.Error(jsonResp)
	//}

	//check time
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		jsonResp := "{\"Error\":\"GetTxTimestamp invalid, Error!!!\"}"
		return shim.Error(jsonResp)
	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		jsonResp := "{\"Error\":\"Vote is over\"}"
		return shim.Error(jsonResp)
	}

	//save support
	indexHistory := make(map[uint64]uint8)
	indexRepeat := false
	for _, oneSupport := range requests {
		topicIndex := oneSupport.TopicIndex - 1
		if _, ok := indexHistory[topicIndex]; ok { //check select repeat
			indexRepeat = true
			break
		}
		indexHistory[topicIndex] = 1
		if topicIndex < uint64(len(topicSupports)) { //1.check index, must not out of total
			if uint64(len(oneSupport.SelectIndexs)) <= topicSupports[topicIndex].SelectMax { //2.check one select's options, must not out of select's max
				lenOfVoteResult := uint64(len(topicSupports[topicIndex].VoteResults))
				selIndexHistory := make(map[uint64]uint8)
				for _, index := range oneSupport.SelectIndexs {
					selectIndex := index - 1
					if _, ok := selIndexHistory[selectIndex]; ok { //check select repeat
						break
					}
					selIndexHistory[selectIndex] = 1
					if selectIndex < lenOfVoteResult { //3.index must be real select options
						topicSupports[topicIndex].VoteResults[selectIndex].Num += voteNum //1
					}
				}
			}
		}
	}
	if indexRepeat {
		jsonResp := "{\"Error\":\"Repeat index of select option \"}"
		return shim.Error(jsonResp)
	}
	voteContentJson, err := json.Marshal(topicSupports)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to generate voteContent Json\"}"
		return shim.Error(jsonResp)
	}
	tkInfo.VoteContent = voteContentJson

	//save token information
	err = setSymbols(stub, tkInfo)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to set symbols\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte(""))
}

type tokenIDInfo struct {
	IsVoteEnd      bool
	CreateAddr     string
	TotalSupply    uint64
	SupportResults []supportResult
	AssetID        string
}
type supportResult struct {
	TopicIndex  uint64
	TopicTitle  string
	VoteResults []voteResult
}

// A slice of TopicResult that implements sort.Interface to sort by Value.
type voteResultList []voteResult

func (p voteResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p voteResultList) Len() int           { return len(p) }
func (p voteResultList) Less(i, j int) bool { return p[i].Num > p[j].Num }

// A function to turn a map into a TopicResultList, then sort and return it.
func sortSupportByCount(tpl voteResultList) voteResultList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}

//GetVoteResult
func (v *Vote) GetVoteResult(stub shim.ChaincodeStubInterface, assetID string) (*tokenIDInfo, error) {
	//assetIDStr
	assetID = strings.ToUpper(assetID)
	//check name is exist or not
	tkInfo := getSymbols(stub, assetID)
	if tkInfo == nil {
		return nil, fmt.Errorf(jsonResp2)
	}

	//get token information
	var topicSupports []topicSupports
	err := json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf(jsonResp1)
	}

	//
	isVoteEnd := false
	headerTime, err := stub.GetTxTimestamp(10)
	if err != nil {
		jsonResp := "{\"Error\":\"GetTxTimestamp invalid, Error!!!\"}"
		return nil, fmt.Errorf(jsonResp)
	}
	if headerTime.Seconds > tkInfo.VoteEndTime.Unix() {
		isVoteEnd = true
	}
	//calculate result
	supportResults := make([]supportResult, 0, len(topicSupports))
	for i, oneTopicSupport := range topicSupports {
		var oneResult supportResult
		oneResult.TopicIndex = uint64(i) + 1
		oneResult.TopicTitle = oneTopicSupport.TopicTitle
		oneResultSort := sortSupportByCount(oneTopicSupport.VoteResults)
		oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort...)
		//for i := uint64(0); i < oneTopicSupport.SelectMax; i++ {
		//	oneResult.VoteResults = append(oneResult.VoteResults, oneResultSort[i])
		//}
		supportResults = append(supportResults, oneResult)
	}

	//token
	asset := tkInfo.AssetID
	tkID := tokenIDInfo{IsVoteEnd: isVoteEnd, CreateAddr: tkInfo.CreateAddr, TotalSupply: tkInfo.TotalSupply,
		SupportResults: supportResults, AssetID: asset.String()}
	return &tkID, nil
}

type voteInfo struct {
	Name        string
	CreateAddr  string
	VoteType    byte
	TotalSupply uint64
	VoteEndTime string
	VoteTopics  []voteTopicIndex
	AssetID     string
}
type voteTopicIndex struct {
	TopicIndex    uint64
	TopicTitle    string
	SelectOptions []string
	SelectMax     uint64
}

//GetVoteInfo
func (v *Vote) GetVoteInfo(stub shim.ChaincodeStubInterface, assetID string) (*voteInfo, error) {
	//assetIDStr
	assetID = strings.ToUpper(assetID)
	//check name is exist or not
	tkInfo := getSymbols(stub, assetID)
	if tkInfo == nil {
		return nil, fmt.Errorf(jsonResp2)
	}

	//get token information
	var topicSupports []topicSupports
	err := json.Unmarshal(tkInfo.VoteContent, &topicSupports)
	if err != nil {
		return nil, fmt.Errorf(jsonResp1)
	}

	//topic info
	voteTopicIndexs := make([]voteTopicIndex, 0, len(topicSupports))
	for i, oneTopicSupport := range topicSupports {
		var oneResult voteTopicIndex
		oneResult.TopicIndex = uint64(i) + 1
		oneResult.TopicTitle = oneTopicSupport.TopicTitle
		for i := range oneTopicSupport.VoteResults {
			oneResult.SelectOptions = append(oneResult.SelectOptions, oneTopicSupport.VoteResults[i].SelectOption)
		}
		oneResult.SelectMax = oneTopicSupport.SelectMax
		voteTopicIndexs = append(voteTopicIndexs, oneResult)
	}

	//token
	asset := tkInfo.AssetID
	tkID := voteInfo{Name: tkInfo.Name, CreateAddr: tkInfo.CreateAddr, VoteType: tkInfo.VoteType,
		TotalSupply: tkInfo.TotalSupply, VoteEndTime: tkInfo.VoteEndTime.String(),
		VoteTopics: voteTopicIndexs, AssetID: asset.String()}
	return &tkID, nil
}
