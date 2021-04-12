package testcli

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	abcicli "github.com/tendermint/classic/abci/client"
	abci "github.com/tendermint/classic/abci/types"
	"github.com/tendermint/classic/crypto/ed25519"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"
)

func StartSocketClient() abcicli.Client {
	client, err := abcicli.NewClient("tcp://127.0.0.1:26658", "socket", true)
	if err != nil {
		panic(err.Error())
	}
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	client.SetLogger(logger.With("module", "abcicli"))
	if err := client.Start(); err != nil {
		panic(fmt.Sprintf("connecting to abci_app: %v", err.Error()))
	}
	return client
}

func InitChain(client abcicli.Client) error {
	total := 10
	vals := make([]abci.ValidatorUpdate, total)
	for i := 0; i < total; i++ {
		pubkey := ed25519.GenPrivKey().PubKey()
		power := int64(cmn.RandInt())
		vals[i] = abci.ValidatorUpdate{pubkey.Address(), pubkey, power}
	}
	_, err := client.InitChainSync(abci.RequestInitChain{
		Validators: vals,
	})
	if err != nil {
		return err
	}
	return nil
}

func SetOption(client abcicli.Client, key, value string) error {
	_, err := client.SetOptionSync(abci.RequestSetOption{Key: key, Value: value})
	if err != nil {
		return err
	}
	return nil
}

func Commit(client abcicli.Client, hashExp []byte) error {
	res, err := client.CommitSync()
	data := res.Data
	if err != nil {
		return err
	}
	if !bytes.Equal(data, hashExp) {
		return errors.New("CommitTx failed")
	}
	return nil
}

func DeliverTx(client abcicli.Client, txBytes []byte, errExp abci.Error, dataExp []byte) error {
	res, _ := client.DeliverTxSync(abci.RequestDeliverTx{Tx: txBytes})
	err, data := res.Error, res.Data
	if errExp == nil {
		if err == nil {
			return nil
		} else {
			return errors.New("DeliverTx error -- expected no error")
		}
	} else {
		if err == nil {
			return errors.New("DeliverTx error -- expected error")
		} else {
			if bytes.Equal(amino.MustMarshalAny(err), amino.MustMarshalAny(errExp)) {
				return nil
			} else {
				errors.New("DeliverTx error -- error mismatch")
			}
		}
	}
	if !bytes.Equal(data, dataExp) {
		return errors.New("DeliverTx error -- data mismatch")
	}
	return nil
}

func CheckTx(client abcicli.Client, txBytes []byte, errExp abci.Error, dataExp []byte) error {
	res, _ := client.CheckTxSync(abci.RequestCheckTx{Tx: txBytes})
	err, data := res.Error, res.Data
	if errExp == nil {
		if err == nil {
			return nil
		} else {
			return errors.New("CheckTx error -- expected no error")
		}
	} else {
		if err == nil {
			return errors.New("CheckTx error -- expected error")
		} else {
			if bytes.Equal(amino.MustMarshalAny(err), amino.MustMarshalAny(errExp)) {
				return nil
			} else {
				errors.New("CheckTx error -- error mismatch")
			}
		}
	}
	if !bytes.Equal(data, dataExp) {
		return errors.New("CheckTx error -- data mismatch")
	}
	return nil
}
