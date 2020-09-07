package manger

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/rwset"
	"github.com/palletone/go-palletone/contracts/scc"
	"github.com/palletone/go-palletone/contracts/ucc"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/vm/common"
	db "github.com/palletone/go-palletone/contracts/comm"
	cclist "github.com/palletone/go-palletone/contracts/list"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	md "github.com/palletone/go-palletone/dag/modules"
)

//type TempCC struct {
//	templateId  []byte
//	name        string
//	path        string
//	version     string
//	description string
//	abi         string
//	language    string
//}

// contract manger module init
func Init(dag dag.IDag, jury core.IAdapterJury) error {
	if err := db.SetCcDagHand(dag); err != nil {
		return err
	}
	if err := peerServerInit(jury); err != nil {
		log.Errorf("peerServerInit:%s", err)
		return err
	}
	if err := systemContractInit(); err != nil {
		log.Errorf("systemContractInit error:%s", err)
		return err
	}
	log.Info("contract manger init ok")

	return nil
}

func InitNoSysCCC(jury core.IAdapterJury) error {
	if err := peerServerInit(jury); err != nil {
		log.Errorf("peerServerInit error:%s", err)
		return err
	}
	return nil
}

func Deinit() error {
	if err := peerServerDeInit(); err != nil {
		log.Errorf("peerServerDeInit error:%s", err)
		return err
	}

	if err := systemContractDeInit(); err != nil {
		log.Errorf("systemContractDeInit error:%s", err)
		return err
	}
	return nil
}

func GetSysCCList() (ccInf []cclist.CCInfo, ccCount int, errs error) {
	scclist := make([]cclist.CCInfo, 0)
	ci := cclist.CCInfo{}

	cclist, count, err := scc.SysCCsList()
	for _, ccinf := range cclist {
		ci.Name = ccinf.Name
		ci.Path = ccinf.Path
		//ci.Enable = ccinf.Enabled
		//ci.SysCC = true
		scclist = append(scclist, ci)
	}
	return scclist, count, err
}

//install but not into db
func Install(dag dag.IDag, chainID, ccName, ccPath, ccVersion, ccLanguage string) (payload *md.ContractTplPayload, err error) {
	log.Info("Install enter", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "cclanguage", ccLanguage)
	defer log.Info("Install exit", "chainID", chainID, "name", ccName, "path", ccPath, "version", ccVersion, "cclanguage", ccLanguage)

	//  用于合约实列
	usrcc := &ucc.UserChaincode{
		Name:     ccName,
		Path:     ccPath,
		Version:  ccVersion,
		Language: ccLanguage,
		Enabled:  true,
	}

	//  产生唯一模板id
	var buffer bytes.Buffer
	buffer.Write([]byte(ccName))
	buffer.Write([]byte(ccPath))
	buffer.Write([]byte(ccVersion))
	tpid := crypto.Keccak256Hash(buffer.Bytes())
	payloadUnit := &md.ContractTplPayload{
		TemplateId: tpid[:],
	}

	//查询一下是否已经安装过
	if tpl, _ := dag.GetContractTpl(tpid[:]); tpl != nil {
		errMsg := fmt.Sprintf("install ,the contractTlp is exist.tplId:%x", tpid)
		log.Debug("Install", "err", errMsg)
		return nil, errors.New(errMsg)
	}

	//将合约代码文件打包成 tar 文件
	paylod, err := ucc.GetUserCCPayload(usrcc)
	if err != nil {
		log.Error("getUserCCPayload err:", "error", err)
		return nil, err
	}
	payloadUnit.ByteCode = paylod

	return payloadUnit, nil
}

func Deploy(rwM rwset.TxManager, idag dag.IDag, chainID string, templateId []byte, txId string, args [][]byte, timeout time.Duration) (deployId []byte, deployPayload *md.ContractDeployPayload, e error) {
	log.Info("Deploy enter", "chainID", chainID, "templateId", templateId, "txId", txId)
	defer log.Info("Deploy exit", "chainID", chainID, "templateId", templateId, "txId", txId)
	setTimeOut := time.Duration(30) * time.Second
	if timeout > 0 {
		setTimeOut = timeout
	}

	// 通过模板id获取源码
	templateCC, chaincodeData, err := ucc.RecoverChainCodeFromDb(idag, templateId)
	if err != nil {
		log.Error("Deploy", "chainid:", chainID, "templateId:", templateId, "RecoverChainCodeFromDb err", err)
		return nil, nil, err
	}

	mksupt := &SupportImpl{}
	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txId)
	if err != nil {
		log.Error("getTxSimulator err:", "error", err)
		return nil, nil, errors.WithMessage(err, "GetTxSimulator error")
	}

	txHash := common.HexToHash(txId)
	depId := crypto.RequestIdToContractAddress(txHash) //common.NewAddress(btxId[:20], common.ContractHash)
	usrccName := depId.String()

	//  TODO 可以开启单机多容器,防止容器名冲突
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[templateCC.Language]),
		Input: &pb.ChaincodeInput{
			Args: args,
		},
		ChaincodeId: &pb.ChaincodeID{
			Name:    usrccName,
			Path:    templateCC.Path,
			Version: templateCC.Version + ":" + contractcfg.GetConfig().ContractAddress,
		},
	}
	//TODO 这里获取运行用户合约容器的相关资源  CpuQuota  CpuShare  MEMORY
	cp := idag.GetChainParameters()
	spec.CpuQuota = cp.UccCpuQuota  //微妙单位（100ms=100000us=上限为1个CPU）
	spec.CpuShare = cp.UccCpuShares //占用率，默认1024，即可占用一个CPU，相对值
	spec.Memory = cp.UccMemory      //字节单位 物理内存  1073741824  1G 2147483648 2G 209715200 200m 104857600 100m
	err = ucc.DeployUserCC(depId.Bytes(), chaincodeData, spec, chainID, txId, txsim, setTimeOut)
	if err != nil {
		log.Error("deployUserCC err:", "error", err)
		return nil, nil, errors.WithMessage(err, "Deploy fail")
	}

	unit, err := RwTxResult2DagDeployUnit(txsim, templateId, usrccName, depId.Bytes(), args, timeout)
	if err != nil {
		log.Errorf("chainID[%s] converRwTxResult2DagUnit failed", chainID)
		return nil, nil, errors.WithMessage(err, "Conver RwSet to dag unit fail")
	}
	return depId.Bytes(), unit, err
}

//func GetChaincode(dag dag.IDag, contractId common.Address) (*cclist.CCInfo, error) {
//	return dag.GetChaincode(contractId)
//}
//
//func SaveChaincode(dag dag.IDag, contractId common.Address, chaincode *cclist.CCInfo) error {
//	return dag.SaveChaincode(contractId, chaincode)
//}
//
//func GetChaincodes(dag dag.IDag) ([]*cclist.CCInfo, error) {
//	return dag.RetrieveChaincodes()
//}

//timeout:ms
// ccName can be contract Id
//func Invoke(chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*peer.ContractInvokePayload, error) {
func Invoke(rwM rwset.TxManager, idag dag.IDag, chainID string, deployId []byte, txid string, args [][]byte, timeout time.Duration) (*md.ContractInvokeResult, error) {
	log.Debugf("Invoke enter")
	log.Info("Invoke enter", "chainID", chainID, "deployId", deployId, "txid", txid, "timeout", timeout)
	defer log.Info("Invoke exit", "chainID", chainID, "deployId", deployId, "txid", txid, "timeout", timeout)

	var mksupt Support = &SupportImpl{}
	creator := []byte("palletone")
	address := common.NewAddress(deployId, common.ContractHash)

	var contractName string
	var contractVersion string
	if address.IsSystemContractAddress() {
		ver := getContractSysVersion(deployId, idag.GetChainParameters().ContractSystemVersion)
		cc, err := cclist.GetChaincode(chainID, deployId, ver)
		if err != nil {
			return nil, err
		}
		contractName = cc.Name
		contractVersion = cc.Version
	} else {
		contract, err := idag.GetContract(address.Bytes())
		if err != nil {
			log.Debugf("Invoke, get chain code err:%s", err.Error())
			return nil, err
		}
		contractName = contract.Name
		contractVersion = contract.Version + ":" + contractcfg.GetConfig().ContractAddress
	}
	startTm := time.Now()
	es := NewEndorserServer(mksupt)
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{Name: contractName, Version: contractVersion},
		//Type:        pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[cc.Language]),
		Input: &pb.ChaincodeInput{Args: args},
	}
	cid := &pb.ChaincodeID{
		//Path:    cc.Path,
		Name:    contractName,
		Version: contractVersion,
	}
	sprop, prop, err := SignedEndorserProposa(chainID, txid, spec, creator, []byte("msg1"))
	if err != nil {
		log.Errorf("signedEndorserProposa error[%v]", err)
		return nil, err
	}
	rsp, unit, err := es.ProcessProposal(rwM, idag, deployId, context.Background(), sprop, prop, chainID, cid, timeout)
	if err != nil {
		log.Infof("ProcessProposal error[%v]", err)
		return nil, err
	}
	if !address.IsSystemContractAddress() {
		sizeRW, disk, isOver := removeConWhenOverDisk(contractName+":"+contractVersion, idag)
		if isOver {
			log.Debugf("utils.KillAndRmWhenOver name = %s,sizeRW = %d,disk = %d", contractName, sizeRW, disk)
			return nil, fmt.Errorf("utils.KillAndRmWhenOver name = %s,sizeRW = %d bytes,disk = %d bytes", contractName, sizeRW, disk)
		}
	}
	stopTm := time.Now()
	duration := stopTm.Sub(startTm)
	//unit.ExecutionTime = duration
	requstId := common.HexToHash(txid)
	unit.RequestId = requstId
	log.Debugf("Invoke Ok, ProcessProposal duration=%v,rsp=%v,%s", duration, rsp, unit.Payload)
	return unit, nil
}

func Stop(rwM rwset.TxManager, idag dag.IDag, contractid []byte, chainID string, txid string, deleteImage bool, dontRmCon bool) (*md.ContractStopPayload, error) {
	log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", contractid, "txid", txid)
	defer log.Info("Stop enter", "contractid", contractid, "chainID", chainID, "deployId", contractid, "txid", txid)

	setChainId := dag.ContractChainId
	if chainID != "" {
		setChainId = chainID
	}
	if txid == "" {
		return nil, errors.New("input param txid is nil")
	}
	address := common.NewAddress(contractid, common.ContractHash)
	contract, err := idag.GetContract(address.Bytes())
	if err != nil {
		log.Debugf("Invoke, get chain code err:%s", err.Error())
		return nil, err
	}
	contract.Version += ":" + contractcfg.GetConfig().ContractAddress
	stopResult, err := StopByName(contractid, setChainId, txid, contract, deleteImage, dontRmCon)
	if err != nil {
		return nil, err
	}
	return stopResult, err
}

func StopByName(contractid []byte, chainID string, txid string, usercc *md.Contract, deleteImage bool, dontRmCon bool) (*md.ContractStopPayload, error) {
	usrcc := &ucc.UserChaincode{
		Name: usercc.Name,
		//Path:     usercc.Path,
		Version: usercc.Version,
		Enabled: true,
		//Language: usercc.Language,
	}
	err := ucc.StopUserCC(contractid, chainID, usrcc, txid, deleteImage, dontRmCon)
	if err != nil {
		errMsg := fmt.Sprintf("StopUserCC err[%s]-[%s]-err[%s]", chainID, usrcc.Name, err)
		return nil, errors.New(errMsg)
	}
	stopResult := &md.ContractStopPayload{
		ContractId: contractid,
	}
	return stopResult, nil
}

func RestartContainer(idag dag.IDag, chainID string, addr common.Address, txId string) ([]byte, error) {
	log.Info("enter RestartContainer", "chainID", chainID, "contract addr", addr.String(), "txId", txId)
	defer log.Info("exit RestartContainer", "txId", txId)
	//setChainId := "palletone"
	setTimeOut := time.Duration(50) * time.Second
	//if chainID != "" {
	//	setChainId = chainID
	//}
	contract, err := idag.GetContract(addr.Bytes())
	if err != nil {
		log.Debugf("Invoke, get chain code err:%s", err.Error())
		return nil, err
	}
	temptpl, err := idag.GetContractTpl(contract.TemplateId)
	if err != nil {
		log.Debugf("get contract template with id = %s, error: %s", contract.TemplateId, err.Error())
		return nil, err
	}
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value[temptpl.Language]),
		Input: &pb.ChaincodeInput{
			Args: [][]byte{},
		},
		ChaincodeId: &pb.ChaincodeID{
			Name:    contract.Name,
			Path:    temptpl.Path,
			Version: contract.Version + ":" + contractcfg.GetConfig().ContractAddress,
		},
	}
	cp := idag.GetChainParameters()
	spec.CpuQuota = cp.UccCpuQuota  //微妙单位（100ms=100000us=上限为1个CPU）
	spec.CpuShare = cp.UccCpuShares //占用率，默认1024，即可占用一个CPU，相对值
	spec.Memory = cp.UccMemory      //字节单位 物理内存  1073741824  1G 2147483648 2G 209715200 200m 104857600 100m
	_, chaincodeData, err := ucc.RecoverChainCodeFromDb(idag, contract.TemplateId)
	if err != nil {
		log.Error("RestartContainer", "chainid:", chainID, "templateId:", contract.TemplateId, "RecoverChainCodeFromDb err", err)
		return nil, err
	}
	err = ucc.DeployUserCC(addr.Bytes(), chaincodeData, spec, chainID, txId, nil, setTimeOut)
	if err != nil {
		log.Error("RestartContainer err:", "error", err)
		return nil, errors.WithMessage(err, "RestartContainer fail")
	}
	return contract.ContractId, err
}

//  调用的时候，若调用完发现磁盘使用超过系统上限，则kill掉并移除
func removeConWhenOverDisk(containerName string, dag dag.IDag) (sizeRW int64, disk int64, isOver bool) {
	log.Debugf("start KillAndRmWhenOver")
	defer log.Debugf("end KillAndRmWhenOver")
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return 0, 0, false
	}
	//  获取所有容器
	allCon, err := client.ListContainers(docker.ListContainersOptions{All: true, Size: true})
	if err != nil {
		log.Debugf("client.ListContainers %s", err.Error())
		return 0, 0, false
	}
	if len(allCon) > 0 {
		//  获取name对应的容器
		containerName = strings.Replace(containerName, ":", "-", -1)
		cp := dag.GetChainParameters()
		for _, c := range allCon {
			if c.Names[0][1:] == containerName && c.SizeRw > cp.UccDisk {
				err := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
				if err != nil {
					log.Debugf("client.RemoveContainer %s", err.Error())
					return 0, 0, false
				}
				log.Debugf("remove container %s", c.Names[0][1:36])
				return c.SizeRw, cp.UccDisk, true
			}
		}
	}
	return 0, 0, false
}

//func StartChaincodeContainer(idag dag.IDag, chainID string, deployId []byte, txId string) ([]byte, error) {
//	//GoStart()
//	return nil, nil
//}

//func DeployByName(rwM rwset.TxManager, idag dag.IDag, chainID string, txid string, ccName string, ccPath string, ccVersion string, args [][]byte, timeout time.Duration) (depllyId []byte, respPayload *md.ContractDeployPayload, e error) {
//	var mksupt Support = &SupportImpl{}
//	setChainId := "palletone"
//	setTimeOut := time.Duration(30) * time.Second
//	if chainID != "" {
//		setChainId = chainID
//	}
//	if timeout > 0 {
//		setTimeOut = timeout
//	}
//	if txid == "" || ccName == "" || ccPath == "" {
//		return nil, nil, errors.New("input param is nil")
//	}
//	randNum, err := crypto.GetRandomNonce()
//	if err != nil {
//		return nil, nil, errors.New("crypto.GetRandomNonce error")
//	}
//	txsim, err := mksupt.GetTxSimulator(rwM, idag, chainID, txid)
//	if err != nil {
//		return nil, nil, errors.New("GetTxSimulator error")
//	}
//	usrcc := &ucc.UserChaincode{
//		Name:     ccName,
//		Path:     ccPath,
//		Version:  ccVersion,
//		InitArgs: args,
//		Enabled:  true,
//	}
//	spec := &pb.ChaincodeSpec{
//		Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]),
//		Input: &pb.ChaincodeInput{
//			Args: args,
//		},
//		ChaincodeId: &pb.ChaincodeID{
//			Name:    ccName,
//			Path:    ccPath,
//			Version: ccVersion,
//		},
//	}
//	err = ucc.DeployUserCC(nil, spec, setChainId, usrcc, txid, txsim, setTimeOut)
//	if err != nil {
//		return nil, nil, errors.New("Deploy fail")
//	}
//	cc := &cclist.CCInfo{
//		Id:      randNum,
//		Name:    ccName,
//		Path:    ccPath,
//		Version: ccVersion,
//		SysCC:   false,
//		//Enable:  true,
//	}
//	err = cclist.SetChaincode(setChainId, 0, cc)
//	if err != nil {
//		log.Errorf("setchaincode[%s]-[%s] fail", setChainId, cc.Name)
//	}
//	return cc.Id, nil, err
//}
