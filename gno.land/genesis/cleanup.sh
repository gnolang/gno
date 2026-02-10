#!/bin/sh

SRC=genesis_txs.jsonl.old
TARGET=genesis_txs.jsonl

> $TARGET

# first, bank transfers, without faucet
cat $SRC | grep /bank.MsgSend | grep -v g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv >> $TARGET

# then, r/demo/users's invites
cat $SRC | grep '"pkg_path":"gno.land/r/demo/users","func":"Invite"' >> $TARGET

# then, r/demo/users's registers
cat $SRC | grep '"pkg_path":"gno.land/r/demo/users","func":"Register"' >> $TARGET

ADMIN_ADDR=g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj
JAE_ADDR=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

# r/boards:gnolang
cat $SRC | grep '"func":"CreateBoard","args":\["gnolang"\]' | sed "s/$JAE_ADDR/$ADMIN_ADDR/" >> $TARGET

# r/boards:gnolang/xxx by jae, admin
cat $SRC | grep "caller\":\"$JAE_ADDR" | grep 'gno.land/r/archive/boards","func":"CreateThread","args":\["1"' | sed "s/$JAE_ADDR/$ADMIN_ADDR/" >> $TARGET
cat $SRC | grep "caller\":\"$ADMIN_ADDR" | grep 'gno.land/r/archive/boards","func":"CreateThread","args":\["1"' >> $TARGET

wc -l $SRC $TARGET
