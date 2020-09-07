package modules

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/stretchr/testify/assert"
)

// The values in those tests are from the Transaction Tests
// at github.com/ethereum/tests.
func TestTransactionHash(t *testing.T) {
	tx := newTestTx()
	t.Log(tx.Hash().String())
	payment := newPaymenForTestt(true)
	msg := NewMessage(APP_PAYMENT, payment)
	tx.AddMessage(msg)
	t.Log(tx.Hash().String())
	payment.LockTime = 1
	assert.NotEqual(t, uint32(1), tx.TxMessages()[3].Payload.(*PaymentPayload).LockTime)
}
func TestTransactionJson(t *testing.T) {
	pay1s := &PaymentPayload{
		LockTime: 12345,
	}
	output := NewTxOut(99999999999999999, []byte{0xee, 0xbb}, NewPTNAsset())
	pay1s.AddTxOut(output)
	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	input := NewTxIn(NewOutPoint(hash, 0, 1), []byte{})
	pay1s.AddTxIn(input)
	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	txmsg := NewTransaction(
		[]*Message{msg},
	)
	data, err := json.Marshal(txmsg)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("json ", string(data))
	}

	tx_item := new(Transaction)
	errNew := json.Unmarshal(data, tx_item)
	if errNew != nil {
		fmt.Println(errNew)
	} else {
		data1, err1 := json.Marshal(tx_item.TxMessages()[0].Payload)
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		fmt.Println("str_date1:", string(data1))
		payment := new(PaymentPayload)
		err2 := json.Unmarshal(data1, &payment)
		if err2 != nil {
			fmt.Println(err2)
		} else {
			fmt.Println("value:=", payment.Outputs[0].Value)
		}
	}

}

func TestTxHash(t *testing.T) {
	tx := newTestTx()

	re, _ := json.Marshal(tx)
	fmt.Println("New Json:", string(re))
	tx2 := &Transaction{}
	err := json.Unmarshal(re, tx2)
	assert.Nil(t, err)
	t.Logf("Unmarsal tx:%#v", tx2)
	assert.Equal(t, tx2.Hash(), tx.Hash())
	msg := tx2.TxMessages()[0]
	for _, in := range msg.Payload.(*PaymentPayload).Inputs {
		in.Extra = []byte("test!")
	}

	tx3 := NewTransaction(tx2.TxMessages())
	for _, input := range tx3.TxMessages()[0].Payload.(*PaymentPayload).Inputs {
		fmt.Println("extra:", string(input.Extra))
	}
	assert.Equal(t, tx3.Hash().String(), tx2.Hash().String())
}
func TestTxClone(t *testing.T) {
	tx := newTestTx()
	tx2 := tx.Clone()

	t.Logf("%#v", tx2)
	t.Logf("msg count:%d", len(tx2.TxMessages()))
	assert.Equal(t, tx.Hash().String(), tx2.Hash().String())
}

func newTestTx() *Transaction {
	pay1s := &PaymentPayload{
		LockTime: 12345,
	}

	output := NewTxOut(Ptn2Dao(10), []byte{0xee, 0xbb}, NewPTNAsset())
	pay1s.AddTxOut(output)

	hash := common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	input := Input{}
	input.PreviousOutPoint = NewOutPoint(hash, 0, 1)
	input.SignatureScript = []byte{}
	input.Extra = []byte("Coinbase")
	fmt.Println(input)
	fmt.Println(input.PreviousOutPoint)
	pay1s.AddTxIn(&input)

	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	msg2 := &Message{
		App:     APP_DATA,
		Payload: &DataPayload{MainData: []byte("Hello PalletOne"), ExtraData: []byte("Hi PalletOne")},
	}

	req := &ContractInvokeRequestPayload{ContractId: []byte{123}, Args: [][]byte{{0x11}, {0x22}}, Timeout: 300}
	msg3 := &Message{App: APP_CONTRACT_INVOKE_REQUEST, Payload: req}
	tx := newTransaction(
		[]*Message{msg, msg2, msg3},
	)
	fmt.Println("payload：", msg.Payload, msg2.Payload, msg3.Payload)
	return tx
}

func TestTransactionEncode(t *testing.T) {

	txmsg3 := newTestTx()
	t.Log("data", txmsg3)
	//emptyTx.SetHash(common.HexToHash("095e7baea6a6c7c4c2dfeb977efac326af552d87"))
	//rightvrsTx.SetHash(common.HexToHash("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"))
	txb, err := rlp.EncodeToBytes(txmsg3)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	//should := common.FromHex("f9010aa0000000000000000000000000b94f5374fce5edbc8e2a8697c15331677e6ebf0bf8e4f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878f84a877061796d656e74a00000000000000000000000000000000000000000000000000000000000000000a07878787878787878787878787878787878787878787878787878787878787878823039")
	//if !bytes.Equal(txb, should) {
	//	log.Error("encoded RLP mismatch", "error", txb)
	//}
	//rlp_hash := new(common.Hash)
	//*rlp_hash = rlp.RlpHash(txmsg3)
	//rightvrsTx.SetHash(*rlp_hash)
	// storage test
	t.Logf("rlp tx: %x", txb)

	//tx := &TestTransaction{}
	tx := &Transaction{}
	err = rlp.DecodeBytes(txb, tx)
	if err != nil {
		t.Error(err)
	}
	t.Log("data", tx)

	assertEqualRlp(t, txmsg3, tx)
	msgs := tx.TxMessages()
	for _, msg := range msgs {
		if msg.App == APP_PAYMENT {
			pay := msg.Payload.(*PaymentPayload)
			fmt.Println("msg", pay.Inputs, pay.Outputs)
			for _, out := range pay.Outputs {
				fmt.Println("info:= ", out)
			}
		}
	}
	if len(msgs) != 3 {
		t.Error("Rlp decode message count error")
	}
	msg0 := msgs[0]
	if msg0.App != APP_PAYMENT {
		t.Error("Payment decode error")
	}
	payment := msg0.Payload.(*PaymentPayload)
	if payment.LockTime != 12345 {
		t.Error("payment locktime decode error.")
	}
	if len(payment.Outputs) == 0 {
		t.Error("payment out decode error.")
	}
	if len(payment.Inputs) == 0 {
		t.Error("payment input decode error.")
	}

	fmt.Printf("PaymentData:%+v", payment)
	msg2 := msgs[1]
	if msg2.App != APP_DATA {
		t.Error("Data decode error")
	}
	data := msg2.Payload.(*DataPayload)
	if len(data.MainData) == 0 {
		t.Error("DataPayload MainData decode error.")
	}
	if len(data.ExtraData) == 0 {
		t.Error("DataPayload ExtraData decode error.")
	}
	t.Log("DataPayload:", data)

	msg3 := msgs[2]
	if msg3.App != APP_CONTRACT_INVOKE_REQUEST {
		t.Error("Data decode error")
	}
	result := msg3.Payload.(*ContractInvokeRequestPayload)
	if result.Timeout == 0 {
		t.Error("ContractInvokeRequestPayload Timeout decode error.")
	}
	if len(result.Args) == 0 {
		t.Error("ContractInvokeRequestPayload Args decode error.")
	}
	if len(result.ContractId) == 0 {
		t.Error("ContractInvokeRequestPayload ContractId decode error.")
	}
	t.Log("ContractInvokeRequestPayload:", result)
}
func TestIDType16Hex(t *testing.T) {
	PTNCOIN := AssetId{'p', 't', 'n', 'c', 'o', 'i', 'n'}
	fmt.Println("ptn hex:", PTNCOIN.String())
	fmt.Println("ptn hex:", PTNCOIN)
	fmt.Println("btc hex:", BTCCOIN.String())
	key := fmt.Sprintf("%s_%s_1_%d", constants.UNIT_HASH_NUMBER_PREFIX, "abc", 100)
	slice := strings.Split(key, fmt.Sprintf("%s_%s_1_", constants.UNIT_HASH_NUMBER_PREFIX, "abc"))
	fmt.Println("result:", len(slice), "0:", slice[0], "1:", slice[1])

	var tx Transaction
	str := "{\"txhash\":\"0xaa0fbe87c07b063cd6a88ab8e2c0075bec35bc80a56956cd50ce98aad3febca6\",\"messages\":[{\"App\":0,\"Payload\":{\"Inputs\":[{\"PreviousOutPoint\":null,\"SignatureScript\":null,\"Extra\":\"W+vkvg==\"}],\"Outputs\":[{\"Value\":100000000,\"PkScript\":\"dqkUj1ulfgUxOae0LG5IueWUIzBQk2WIrA==\",\"Asset\":{\"asset_id\":[119,169,59,162,215,104,17,232,157,4,140,133,144,10,158,67],\"unique_id\":[119,169,59,162,215,104,17,232,157,4,140,133,144,10,158,67],\"chain_id\":1}}],\"LockTime\":0}}]}"
	err := json.Unmarshal([]byte(str), &tx)
	fmt.Println("error: ", err)
	for _, msg := range tx.TxMessages() {
		fmt.Println("info: ", msg.Payload)
		data, err := json.Marshal(msg.Payload)
		if err != nil {
			return
		}
		payment := new(PaymentPayload)
		err1 := json.Unmarshal(data, &payment)
		for j, out := range payment.Outputs {
			fmt.Println("payment: ", err1, j, out)
		}

	}
}
func TestTransaction_EncodeRLP_Size(t *testing.T) {
	pay1s := PaymentPayload{
		LockTime: 12345,
	}
	a := &Asset{AssetId: PTNCOIN}

	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay1s.AddTxOut(output)

	msg := &Message{
		App:     APP_PAYMENT,
		Payload: pay1s,
	}
	msgs := make([]*Message, 0)
	for i := 1; i < 1000; i++ {
		msgs = append(msgs, msg)
		tx := NewTransaction(msgs)
		txb, _ := rlp.EncodeToBytes(tx)
		t.Logf("input count:{%d}, encode tx size:%d\n", i, len(txb))
	}
}

func TestRlpdecodeValue(t *testing.T) {
	str1 := "[{\"address\":\"P1DU7BHzyVU3eehHKySqgeEZhZC8oQo1yaM\",\"content\":\"{\\\"key\\\",\\\"value\\\"}\",\"time\":858877}]"

	strBytes, _ := rlp.EncodeToBytes(str1)
	fmt.Println(strBytes)

	var val []byte
	err := rlp.DecodeBytes(strBytes, &val)
	fmt.Println("error", err, "val", string(val))

	str := hexutil.Encode(strBytes)
	fmt.Println(str)
}

func TestPaymentpayloadInputRlp(t *testing.T) {
	i := NewTxIn(nil, []byte("a"))
	b, err := rlp.EncodeToBytes(i)
	assert.Nil(t, err)
	t.Logf("rlp:%x", b)
	i2 := &Input{}
	err = rlp.DecodeBytes(b, i2)
	assert.Nil(t, err)
	assert.Equal(t, i2.SignatureScript, []byte("a"))
}

func TestTransaction_GetTxFee(t *testing.T) {
	tx := newTestTx()
	utxoQueryFn := func(outpoint *OutPoint) (*Utxo, error) {
		t := time.Now().AddDate(0, 0, -1).Unix()
		return &Utxo{Amount: Ptn2Dao(11), Timestamp: uint64(t), Asset: NewPTNAsset()}, nil
	}
	fee, err := tx.GetTxFee(utxoQueryFn)
	assert.Nil(t, err)
	assert.True(t, fee.Amount == Ptn2Dao(1))
	t.Log(fee.String())
	fee2, err := tx.GetTxFee(utxoQueryFn)
	assert.Nil(t, err)
	t.Log(fee2.String())
	assert.Equal(t, fee2.Amount, Ptn2Dao(1))
}
func Ptn2Dao(ptn uint64) uint64 {
	return ptn * 100000000
}

func newPaymenForTestt(includeCoinbase bool) *PaymentPayload {
	pay := &PaymentPayload{LockTime: 123, Inputs: []*Input{}, Outputs: []*Output{}}
	if includeCoinbase {
		pay.Inputs = append(pay.Inputs, &Input{SignatureScript: []byte("test"), Extra: []byte("Extra")})
	}
	hash := common.HexToHash("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac")
	input := &Input{SignatureScript: []byte{1, 2, 3}, Extra: nil, PreviousOutPoint: NewOutPoint(hash, 123, 9999)}
	pay.Inputs = append(pay.Inputs, input)
	a := &Asset{AssetId: PTNCOIN}

	output := NewTxOut(1, common.Hex2Bytes("0x76a914bd05274d98bb768c0e87a55d9a6024f76beb462a88ac"), a)
	pay.Outputs = append(pay.Outputs, output)
	return pay
}

func TestAdditionJson(t *testing.T) {
	income := &Addition{
		Addr:   common.DestroyAddress,
		Amount: 12300000,
		Asset:  NewPTNAsset(),
	}
	data, _ := json.Marshal(income)
	t.Log(string(data))
}
