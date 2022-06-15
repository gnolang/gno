#!/bin/sh

SRC=genesis_txs.txt.old
TARGET=genesis_txs.txt


> $TARGET
# first, bank transfers, without faucet
cat $SRC | grep /bank.MsgSend | grep -v g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv >> $TARGET

wc -l $SRC $TARGET
