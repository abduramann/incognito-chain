package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	_ "github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertOutputCoinToInputCoin(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx := &Tx{}
	err = tx.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}

	in := ConvertOutputCoinToInputCoin(tx.Proof.GetOutputCoins())
	assert.Equal(t, 1, len(in))
	assert.Equal(t, tx.Proof.GetOutputCoins()[0].CoinDetails.GetValue(), in[0].CoinDetails.GetValue())
}

func TestEstimateTxSize(t *testing.T) {
	key, err := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	assert.Equal(t, nil, err)
	err = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	assert.Equal(t, nil, err)
	paymentAddress := key.KeySet.PaymentAddress
	tx := &Tx{}
	err = tx.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}

	payments := []*privacy.PaymentInfo{&privacy.PaymentInfo{
		PaymentAddress: paymentAddress,
		Amount:         5,
	}}

	size := EstimateTxSize(NewEstimateTxSizeParam(tx.Proof.GetOutputCoins(), payments, true, nil, nil, nil, 1))
	assert.Greater(t, size, uint64(0))

	customTokenParams := CustomTokenParamTx{
		Receiver: []TxTokenVout{{PaymentAddress: paymentAddress, Value: 5}},
		vins:     []TxTokenVin{{PaymentAddress: paymentAddress, VoutIndex: 1}},
	}
	size1 := EstimateTxSize(NewEstimateTxSizeParam(tx.Proof.GetOutputCoins(), payments, true, nil, &customTokenParams, nil, 1))
	assert.Greater(t, size1, uint64(0))

	privacyCustomTokenParams := CustomTokenPrivacyParamTx{
		Receiver: []*privacy.PaymentInfo{{
			PaymentAddress: paymentAddress, Amount: 5,
		}},
	}
	size2 := EstimateTxSize(NewEstimateTxSizeParam(tx.Proof.GetOutputCoins(), payments, true, nil, nil, &privacyCustomTokenParams, 1))
	assert.Greater(t, size2, uint64(0))
}

func TestRandomCommitmentsProcess(t *testing.T) {
	key, _ := wallet.Base58CheckDeserialize("112t8rnXCqbbNYBquntyd6EvDT4WiDDQw84ZSRDKmazkqrzi6w8rWyCVt7QEZgAiYAV4vhJiX7V9MCfuj4hGLoDN7wdU1LoWGEFpLs59X7K3")
	_ = key.KeySet.InitFromPrivateKey(&key.KeySet.PrivateKey)
	paymentAddress := key.KeySet.PaymentAddress
	tx1 := &Tx{}
	err := tx1.InitTxSalary(10, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}
	db.StoreCommitments(common.Hash{}, paymentAddress.Pk, [][]byte{tx1.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().Compress()}, 0)

	in1 := ConvertOutputCoinToInputCoin(tx1.Proof.GetOutputCoins())

	cmmIndexs, myIndexs, cmm := RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in1, 0, db, 0, &common.Hash{}))
	assert.Equal(t, 8, len(cmmIndexs))
	assert.Equal(t, 1, len(myIndexs))
	assert.Equal(t, 8, len(cmm))

	tx2 := &Tx{}
	err = tx2.InitTxSalary(5, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	if err != nil {
		t.Error(err)
	}
	db.StoreCommitments(common.Hash{}, paymentAddress.Pk, [][]byte{tx2.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().Compress()}, 0)
	tx3 := &Tx{}
	err = tx3.InitTxSalary(5, &paymentAddress, &key.KeySet.PrivateKey, db, nil)
	db.StoreCommitments(common.Hash{}, paymentAddress.Pk, [][]byte{tx3.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment().Compress()}, 0)
	in2 := ConvertOutputCoinToInputCoin(tx2.Proof.GetOutputCoins())
	in := append(in1, in2...)

	cmmIndexs, myIndexs, cmm = RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in, 0, db, 0, &common.Hash{}))
	assert.Equal(t, 16, len(cmmIndexs))
	assert.Equal(t, 16, len(cmm))
	assert.Equal(t, 2, len(myIndexs))

	db.CleanCommitments()
	cmmIndexs1, myCommIndex1, cmm1 := RandomCommitmentsProcess(NewRandomCommitmentsProcessParam(in, 0, db, 0, &common.Hash{}))
	assert.Equal(t, 0, len(cmmIndexs1))
	assert.Equal(t, 0, len(myCommIndex1))
	assert.Equal(t, 0, len(cmm1))
}
