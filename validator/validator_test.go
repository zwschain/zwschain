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
	"encoding/hex"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/errors"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/parameter"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidateUnitTxs(t *testing.T) {
	parameter.CurrentSysParameters.GenerateUnitReward = 0
	//构造一个Unit包含3个Txs，
	//0是Coinbase，收集30000Dao手续费
	//1是普通Tx，20000->15000 付5000Dao手续费，产生1Utxo
	//2是普通Tx， 15000->5000 付10000Dao手续费，使用Tx1中的一个Utxo，产生1Utxo
	tx0 := newCoinbaseTx()
	tx1 := newTx1(t)
	outPoint := modules.NewOutPoint(tx1.Hash(), 0, 0)
	tx2 := newTx2(t, outPoint)
	txs := modules.Transactions{tx0, tx1, tx2}
	dagq := &mockiDagQuery{}
	utxoQuery := &mockUtxoQuery{}
	mockStatedbQuery := &mockStatedbQuery{}
	prop := &mockiPropQuery{}
	validate := NewValidate(dagq, utxoQuery, mockStatedbQuery, prop, newCache(), false)
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	code := validate.validateTransactions(txs, 1564675200, addr)
	assert.Equal(t, code, TxValidationCode_VALID)
}

type mockStatedbQuery struct {
}

func (q *mockStatedbQuery) GetContractTpl(tplId []byte) (*modules.ContractTemplate, error) {
	return nil, nil
}
func (q *mockStatedbQuery) GetMediators() map[common.Address]bool {
	return nil
}

func (q *mockStatedbQuery) GetMediator(add common.Address) *core.Mediator {
	return nil
}
func (q *mockStatedbQuery) GetBlacklistAddress() ([]common.Address, *modules.StateVersion, error) {
	return []common.Address{}, nil, nil
}

func (q *mockStatedbQuery) GetContractJury(contractId []byte) (*modules.ElectionNode, error) {
	return nil, nil
}

func (q *mockStatedbQuery) GetContractState(id []byte, field string) ([]byte, *modules.StateVersion, error) {
	return nil, nil, nil
}

func (q *mockStatedbQuery) GetContractStatesByPrefix(id []byte, prefix string) (map[string]*modules.ContractStateValue, error) {
	return map[string]*modules.ContractStateValue{}, nil
}

func (q *mockStatedbQuery) GetJurorByAddrHash(addrHash common.Hash) (*modules.JurorDeposit, error) {
	return nil, nil
}

func (q *mockStatedbQuery) GetJurorReward(jurorAdd common.Address) common.Address {
	return jurorAdd
}

func (q *mockStatedbQuery) GetTxRequesterAddress(tx *modules.Transaction) (common.Address, error) {
	return common.Address{}, nil
}

func (q *mockStatedbQuery) IsContractDeveloper(addr common.Address) bool {
	return true
}

type mockUtxoQuery struct {
}

func (q *mockUtxoQuery) GetStxoEntry(outpoint *modules.OutPoint) (*modules.Stxo, error) {
	return nil, nil
}

func (q *mockUtxoQuery) GetUtxoEntry(outpoint *modules.OutPoint) (*modules.Utxo, error) {
	hash := common.HexToHash("1")
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	utxo := &modules.Utxo{Amount: 20000, LockTime: 0, Asset: modules.NewPTNAsset(), PkScript: lockScript}
	if outpoint.TxHash == hash {
		return utxo, nil
	}

	return nil, errors.New("No utxo found")
}

func newCoinbaseTx() *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	output := modules.NewTxOut(30000, lockScript, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	input := modules.Input{}
	input.Extra = []byte("Coinbase")

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	return tx
}

func newTx1(t *testing.T) *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	output := modules.NewTxOut(15000, lockScript, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("1")
	input := modules.Input{}
	input.PreviousOutPoint = modules.NewOutPoint(hash, 0, 0)
	input.SignatureScript = []byte{}

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	//Sign

	lockScripts := map[modules.OutPoint][]byte{
		*input.PreviousOutPoint: lockScript,
	}
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, msg)
	}
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages()[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)

	return tx
}
func newTx2(t *testing.T, outpoint *modules.OutPoint) *modules.Transaction {
	pay1s := &modules.PaymentPayload{}
	output := modules.NewTxOut(5000, []byte{}, modules.NewPTNAsset())
	pay1s.AddTxOut(output)
	input := modules.Input{}
	input.PreviousOutPoint = outpoint
	input.SignatureScript = []byte{}

	pay1s.AddTxIn(&input)

	msg := &modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: pay1s,
	}

	tx := modules.NewTransaction(
		[]*modules.Message{msg},
	)
	//Sign
	addr, _ := common.StringToAddress("P1HXNZReTByQHgWQNGMXotMyTkMG9XeEQfX")
	lockScript := tokenengine.Instance.GenerateLockScript(addr)
	lockScripts := map[modules.OutPoint][]byte{
		*input.PreviousOutPoint: lockScript,
	}
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, msg)
	}
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages()[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	return tx
}
func newHeader(txs modules.Transactions) *modules.Header {
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey, _ := hex.DecodeString("038cc8c907b29a58b00f8c2590303bfc93c69d773b9da204337678865ee0cafadb")
	//addr:= crypto.PubkeyBytesToAddress(pubKey)

	b := []byte{}
	header := modules.NewHeader([]common.Hash{hash}, core.DeriveSha(txs), b, b, b, b, []uint16{},
		modules.NewPTNIdType(), 1, int64(15987666666))
	headerHash := header.HashWithoutAuthor()
	sign, _ := crypto.Sign(headerHash[:], privKey)
	header.SetAuthor(modules.Authentifier{PubKey: pubKey, Signature: sign})
	return header
}
func TestValidate_ValidateHeader(t *testing.T) {
	tx := newTx1(t)

	header := newHeader(modules.Transactions{tx})
	stateQ := &mockStatedbQuery{}
	v := NewValidate(nil, nil, stateQ, nil, newCache(), true)
	vresult := v.validateHeaderExceptGroupSig(header, false)
	t.Log(vresult)
	assert.Equal(t, vresult, TxValidationCode_VALID)
}

func TestSignAndVerifyATx(t *testing.T) {

	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")

	pubKeyBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privKeyBytes)
	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyBytesToAddress(pubKeyBytes)
	t.Logf("Addr:%s", addr.String())
	lockScript := tokenengine.Instance.GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	payment.AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	payment2 := &modules.PaymentPayload{}
	utxoTxId2 := common.HexToHash("1651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint2 := modules.NewOutPoint(utxoTxId2, 1, 1)
	txIn2 := modules.NewTxIn(outPoint2, []byte{})
	payment2.AddTxIn(txIn2)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	payment2.AddTxOut(modules.NewTxOut(1, lockScript, asset1))
	m1 := modules.NewMessage(modules.APP_PAYMENT, payment)
	m2 := modules.NewMessage(modules.APP_PAYMENT, payment2)

	m3 := modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("Hello PalletOne")})
	tx := modules.NewTransaction([]*modules.Message{m1, m2, m3})
	lockScripts := map[modules.OutPoint][]byte{
		*outPoint:  lockScript[:],
		*outPoint2: tokenengine.Instance.GenerateP2PKHLockScript(pubKeyHash),
	}
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	addr: privKey,
	//}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return pubKeyBytes, nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		return crypto.MyCryptoLib.Sign(privKeyBytes, hash)
	}
	var hashtype uint32
	hashtype = 1
	_, err := tokenengine.Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages()[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)

}
func TestTime(t *testing.T) {
	ti, _ := time.ParseInLocation("2006-01-02 15:04:05", "2020-03-01 00:00:00", time.Local)
	t.Log(ti.Format("2006-01-02 15:04:05"))
	t.Log(ti.Unix())
	t2 := time.Unix(1570870800, 0)
	t.Logf("Time2:%v", t2.Format("2006-01-02 15:04:05"))
}
