#!/bin/sh

SRC=genesis_txs.txt.old
TARGET=genesis_txs.txt


> $TARGET

# first, bank transfers, without faucet
cat $SRC | grep /bank.MsgSend | grep -v g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv >> $TARGET

# then, r/users's invites
cat $SRC | grep '"pkg_path":"gno.land/r/users","func":"Invite"' >> $TARGET

# then, r/users's registers
cat $SRC | grep '"pkg_path":"gno.land/r/users","func":"Register"' >> $TARGET

# gnolang board
cat $SRC | grep '"func":"CreateBoard","args":\["gnolang"\]' >> $TARGET

wc -l $SRC $TARGET
