package tendermint_test

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
)

func TestVoteBytesToSign(t *testing.T) {
	t.Run("update client 1", func(t *testing.T) {
		var (
			// ref: https://www.mintscan.io/atomone/tx/E7A075A6B6BC56ED2006A91F7833354C7F721CC9618BF09F276B4F58B119149B?sector=json
			// (truncated with only the first 2 validators)
			//	{
			//  "@type": "/ibc.core.client.v1.MsgUpdateClient",
			//  "client_id": "07-tendermint-2",
			//  "client_message": {
			//    "@type": "/ibc.lightclients.tendermint.v1.Header",
			//    "signed_header": {
			//      "header": {
			//        "version": {
			//          "block": "11",
			//          "app": "0"
			//        },
			//        "chain_id": "osmosis-1",
			//        "height": "44873552",
			//        "time": "2025-09-25T07:55:56.299553312Z",
			//        "last_block_id": {
			//          "hash": "rV4iVJhw1gwx1d7nij1tiI0211CsYbqzrlT1bAKB6FM=",
			//          "part_set_header": {
			//            "total": 1,
			//            "hash": "TmTAPEknTC5O3/XV12onhZvcvLiK0X6CdPFND4NMARw="
			//          }
			//        },
			//        "last_commit_hash": "I18uOE4JQq7ZpkGb5oZQqYxLSWO86nFZ8yEICfs+XIU=",
			//        "data_hash": "Wv50wpNbp0d8P7bnNPiDqKOO8fo2RCvs8H8LI2c5dbo=",
			//        "validators_hash": "ehQCurs4+UgURefaP7Sw1T/MnviV6jCE8SLFMGB2Mis=",
			//        "next_validators_hash": "ehQCurs4+UgURefaP7Sw1T/MnviV6jCE8SLFMGB2Mis=",
			//        "consensus_hash": "JIS7qLmeUx4aP2QiNi59PuuZqDJRGZaKNV8FwJ7t1As=",
			//        "app_hash": "DAHBO3lYaIFu9XmplThmF7ijXG7YwDZawGSDKKCHh7c=",
			//        "last_results_hash": "aSOQCy7TAjnW/YXTu8OJjzKmnrLA3loaf0i0y+D9U/Q=",
			//        "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			//        "proposer_address": "ZraWZuv3dufry+GXq6RmpxLicHY="
			//      },
			//      "commit": {
			//        "height": "44873552",
			//        "round": 0,
			//        "block_id": {
			//          "hash": "NpiImIJoaSaIucwNs5cqpgMsL/8wxEPYC3P0jA5aQSI=",
			//          "part_set_header": {
			//            "total": 1,
			//            "hash": "QqzwnLzvixIcUz+hPeUQjDV6NaLkFRKXACCxJIrBHzw="
			//          }
			//        },
			//        "signatures": [
			//          {
			//            "block_id_flag": "BLOCK_ID_FLAG_COMMIT",
			//            "validator_address": "y1pjuR6PTujbk1lCy+JXJGNkeeA=",
			//            "timestamp": "2025-09-25T07:55:57.306746166Z",
			//            "signature": "qtv1z4S2Q6T87vGQo0lrjRZqv9PrHIji4pTyviMnVyGx9td6eySdzwQwCthwmihU48ebNlFiMlFJ0CT891UmDg=="
			//          },
			//          {
			//            "block_id_flag": "BLOCK_ID_FLAG_COMMIT",
			//            "validator_address": "ZraWZuv3dufry+GXq6RmpxLicHY=",
			//            "timestamp": "2025-09-25T07:55:57.310583641Z",
			//            "signature": "Q5E6Kjma00n/T98rC9qJmoB6JTGFX/IB+mDVs4Wd1h0eJ8fabY/6oI8zdoU6/7W6VR6wjpHyWBsJrpGT6C0LCg=="
			//          }
			//        ]
			//      }
			//    },
			//    "validator_set": {
			//      "validators": [
			//        {
			//          "address": "y1pjuR6PTujbk1lCy+JXJGNkeeA=",
			//          "pub_key": {
			//            "ed25519": "6Nz09YGHzwWxjczG0IhK4Iv0qY2IcX0P/5KitvRXTUc="
			//          },
			//          "voting_power": "19087034",
			//          "proposer_priority": "0"
			//        },
			//        {
			//          "address": "ZraWZuv3dufry+GXq6RmpxLicHY=",
			//          "pub_key": {
			//            "ed25519": "wB25StLxbzmD0uTiFiH6xySZd0H13kyanNUvvlUpa34="
			//          },
			//          "voting_power": "11979324",
			//          "proposer_priority": "0"
			//        }
			//      ],
			//      "proposer": {
			//        "address": "ZraWZuv3dufry+GXq6RmpxLicHY=",
			//        "pub_key": {
			//          "ed25519": "wB25StLxbzmD0uTiFiH6xySZd0H13kyanNUvvlUpa34="
			//        },
			//        "voting_power": "11979324",
			//        "proposer_priority": "0"
			//      },
			//      "total_voting_power": "276866533"
			//    },
			//    "trusted_height": {
			//      "revision_number": "1",
			//      "revision_height": "44873531"
			//    },
			//    "trusted_validators": {
			//      "validators": [
			//        {
			//          "address": "y1pjuR6PTujbk1lCy+JXJGNkeeA=",
			//          "pub_key": {
			//            "ed25519": "6Nz09YGHzwWxjczG0IhK4Iv0qY2IcX0P/5KitvRXTUc="
			//          },
			//          "voting_power": "19087034",
			//          "proposer_priority": "0"
			//        },
			//        {
			//          "address": "ZraWZuv3dufry+GXq6RmpxLicHY=",
			//          "pub_key": {
			//            "ed25519": "wB25StLxbzmD0uTiFiH6xySZd0H13kyanNUvvlUpa34="
			//          },
			//          "voting_power": "11979324",
			//          "proposer_priority": "0"
			//        }
			//      ],
			//      "proposer": {
			//        "address": "xx7zwESZhHO0zsyumXcZdvGY1KU=",
			//        "pub_key": {
			//          "ed25519": "c28EquE3UMuallczX/Bu1UHpV9dM/yTaIVIR+ftY1g4="
			//        },
			//        "voting_power": "9284967",
			//        "proposer_priority": "0"
			//      },
			//      "total_voting_power": "276866533"
			//    }
			//  },
			//  "signer": "atone14d0ka5jcuhf9v3n7qpdpmc9qhgsd6uae2g7jxy"
			//}
			chainID       = "osmosis-1"
			timestamp     = toTime("2025-09-25T07:55:56.299553312Z")
			apphash       = b64Dec("DAHBO3lYaIFu9XmplThmF7ijXG7YwDZawGSDKKCHh7c=")
			height        = uint64(44873552)
			blockhash     = b64Dec("NpiImIJoaSaIucwNs5cqpgMsL/8wxEPYC3P0jA5aQSI=")
			parsethash    = b64Dec("QqzwnLzvixIcUz+hPeUQjDV6NaLkFRKXACCxJIrBHzw=")
			consensushash = b64Dec("JIS7qLmeUx4aP2QiNi59PuuZqDJRGZaKNV8FwJ7t1As=")
			val1          = tendermint.NewValidator(
				"y1pjuR6PTujbk1lCy+JXJGNkeeA=", "6Nz09YGHzwWxjczG0IhK4Iv0qY2IcX0P/5KitvRXTUc=", 2,
			)
			val2 = tendermint.NewValidator(
				"ZraWZuv3dufry+GXq6RmpxLicHY=", "wB25StLxbzmD0uTiFiH6xySZd0H13kyanNUvvlUpa34=", 3,
			)
			valset        = tendermint.NewValset(val1, val2)
			trustedValset = tendermint.NewValset(val1, val2)
			signatures    = []tendermint.CommitSig{
				{
					BlockIDFlag:      tendermint.BlockIDFlagCommit,
					ValidatorAddress: valset.Validators[0].Address,
					Timestamp:        toTime("2025-09-25T07:55:57.306746166Z"),
					Signature:        b64Dec("qtv1z4S2Q6T87vGQo0lrjRZqv9PrHIji4pTyviMnVyGx9td6eySdzwQwCthwmihU48ebNlFiMlFJ0CT891UmDg=="),
				},
				{
					BlockIDFlag:      tendermint.BlockIDFlagCommit,
					ValidatorAddress: valset.Validators[1].Address,
					Timestamp:        toTime("2025-09-25T07:55:57.310583641Z"),
					Signature:        b64Dec("Q5E6Kjma00n/T98rC9qJmoB6JTGFX/IB+mDVs4Wd1h0eJ8fabY/6oI8zdoU6/7W6VR6wjpHyWBsJrpGT6C0LCg=="),
				},
			}

			msg = tendermint.NewMsgHeader(
				chainID, timestamp, apphash, blockhash, parsethash, consensushash,
				height, types.Height{}, valset, trustedValset, signatures,
			)

			// expectedBytesToSign are computed from cometbft/types.Commit.VoteSignBytes() function
			expectedBytesToSign = []string{
				"6e08021150b7ac020000000022480a20369888988268692688b9cc0db3972aa6032c2fff30c443d80b73f48c0e5a412212240801122042acf09cbcef8b121c533fa13de5108c357a35a2e41512970020b1248ac11f3c2a0c088debd3c60610b6a6a2920132096f736d6f7369732d31",
				"6e08021150b7ac020000000022480a20369888988268692688b9cc0db3972aa6032c2fff30c443d80b73f48c0e5a412212240801122042acf09cbcef8b121c533fa13de5108c357a35a2e41512970020b1248ac11f3c2a0c088debd3c60610d9c28c940132096f736d6f7369732d31",
			}
		)

		testVotesBytesToSign(t, msg, expectedBytesToSign)
	})

	t.Run("update client 2", func(t *testing.T) {
		var (
			// Ref: https://www.mintscan.io/atomone/tx/DAE6D01666864DBB1F7AFDDD2C20B0C22D34974E9682B15A5049C24574EBEC70?sector=json
			// (truncated with the 3 first validators)
			//{
			//  "@type": "/ibc.core.client.v1.MsgUpdateClient",
			//  "client_id": "07-tendermint-1",
			//  "client_message": {
			//    "@type": "/ibc.lightclients.tendermint.v1.Header",
			//    "signed_header": {
			//      "header": {
			//        "version": {
			//          "block": "11",
			//          "app": "0"
			//        },
			//        "chain_id": "beezee-1",
			//        "height": "19282029",
			//        "time": "2025-09-25T14:00:43.469479971Z",
			//        "last_block_id": {
			//          "hash": "cx4wWNgIMOrgnXOPFlXMTV6je4GbYZjRhgcJ7RNs0g4=",
			//          "part_set_header": {
			//            "total": 1,
			//            "hash": "Nv1TMWmm4ze2Y1p3pBLF9VOPJaPOARp4ypZglD4HZmo="
			//          }
			//        },
			//        "last_commit_hash": "zNO6+/eHsl8hih5kYSEDQql/vZUqTyuc6qJD6AnjrcI=",
			//        "data_hash": "1EbR1y5AQgdELcRJTPf7ql6BJI5J9Ji9mTkdNXElcj4=",
			//        "validators_hash": "ovLukv8mGP7nzYUiYBB8gUWQMvPZehcxQi86HsRGRW4=",
			//        "next_validators_hash": "ovLukv8mGP7nzYUiYBB8gUWQMvPZehcxQi86HsRGRW4=",
			//        "consensus_hash": "BOEs4QnQgrOaBfgH2qUBI3M1r0rPLx++gYSRGuYVAns=",
			//        "app_hash": "XI9E+rmDbjwfn/EvQVp7CAqygaEI8Fingdh3xR+7aKw=",
			//        "last_results_hash": "pgDffcPvKUJK3G2+adQmiLdiY5Rh7jOtRx81ytLcq4k=",
			//        "evidence_hash": "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
			//        "proposer_address": "ysbGRHeOh4hD9fUbkM1RHP9bVfc="
			//      },
			//      "commit": {
			//        "height": "19282029",
			//        "round": 0,
			//        "block_id": {
			//          "hash": "q82PBaLn4Vg3oLm0v670BOm1Ay+CR69DyBqNA/V//l4=",
			//          "part_set_header": {
			//            "total": 1,
			//            "hash": "PqBqLFCb6YhCQZFDnOeHzp1kHAbKs7fwHqOwnn8ha2s="
			//          }
			//        },
			//        "signatures": [
			//          {
			//            "block_id_flag": "BLOCK_ID_FLAG_ABSENT",
			//            "validator_address": null,
			//            "timestamp": "0001-01-01T00:00:00Z",
			//            "signature": null
			//          },
			//          {
			//            "block_id_flag": "BLOCK_ID_FLAG_COMMIT",
			//            "validator_address": "mYLjIF9AxwR+BBlkD9M3f+xqKNE=",
			//            "timestamp": "2025-09-25T14:00:49.373544902Z",
			//            "signature": "rgUFLD3Yk9JdcfWSqcscXeUZVY+Q7E1jQ1Eumkh/LT2fUI4dsk2NBeOhjyKaF5Ey/q6DwoINss6iSxNPj8s2Ag=="
			//          },
			//          {
			//            "block_id_flag": "BLOCK_ID_FLAG_COMMIT",
			//            "validator_address": "ysbGRHeOh4hD9fUbkM1RHP9bVfc=",
			//            "timestamp": "2025-09-25T14:00:49.396076476Z",
			//            "signature": "thHpLNT6555JHoNr10iBpp8NJemRMPKMwnV4QDL6gLpaYChqrQND+8K4hQrC8Ld2IAuIT92ZHFWOmoVH9bIcBw=="
			//          }
			//        ]
			//      }
			//    },
			//    "validator_set": {
			//      "validators": [
			//        {
			//          "address": "h9mY0OvyfYUX7rMfkXO+9bpEWdo=",
			//          "pub_key": {
			//            "ed25519": "wF3Daj7tqX8nVOJd2WCkMVvqyx/lPzuI5/y3wwSRwbY="
			//          },
			//          "voting_power": "7870904",
			//          "proposer_priority": "0"
			//        },
			//        {
			//          "address": "mYLjIF9AxwR+BBlkD9M3f+xqKNE=",
			//          "pub_key": {
			//            "ed25519": "dzCYrLu2sjpWSiEd2MVqsr+Q6ocBDUUUrRKBehGOeLM="
			//          },
			//          "voting_power": "7847038",
			//          "proposer_priority": "0"
			//        },
			//		    {
			// 		      "address": "ysbGRHeOh4hD9fUbkM1RHP9bVfc=",
			// 		      "pub_key": {
			// 		        "ed25519": "tYyou2DP3JWDrvKPtbAZpTCEtzqr9h1nsi/srNZwLiY="
			// 		      },
			// 		      "voting_power": "7381171",
			// 		      "proposer_priority": "0"
			// 		    }
			//      ],
			//      "proposer": {
			//        "address": "ysbGRHeOh4hD9fUbkM1RHP9bVfc=",
			//        "pub_key": {
			//          "ed25519": "tYyou2DP3JWDrvKPtbAZpTCEtzqr9h1nsi/srNZwLiY="
			//        },
			//        "voting_power": "7381171",
			//        "proposer_priority": "0"
			//      },
			//      "total_voting_power": "169513553"
			//    },
			//    "trusted_height": {
			//      "revision_number": "1",
			//      "revision_height": "19280755"
			//    },
			//    "trusted_validators": {
			//      "validators": [
			//        {
			//          "address": "h9mY0OvyfYUX7rMfkXO+9bpEWdo=",
			//          "pub_key": {
			//            "ed25519": "wF3Daj7tqX8nVOJd2WCkMVvqyx/lPzuI5/y3wwSRwbY="
			//          },
			//          "voting_power": "7870904",
			//          "proposer_priority": "0"
			//        },
			//        {
			//          "address": "mYLjIF9AxwR+BBlkD9M3f+xqKNE=",
			//          "pub_key": {
			//            "ed25519": "dzCYrLu2sjpWSiEd2MVqsr+Q6ocBDUUUrRKBehGOeLM="
			//          },
			//          "voting_power": "7847038",
			//          "proposer_priority": "0"
			//        },
			//		    {
			// 		      "address": "ysbGRHeOh4hD9fUbkM1RHP9bVfc=",
			// 		      "pub_key": {
			// 		        "ed25519": "tYyou2DP3JWDrvKPtbAZpTCEtzqr9h1nsi/srNZwLiY="
			// 		      },
			// 		      "voting_power": "7381171",
			// 		      "proposer_priority": "0"
			// 		    }
			//      ],
			//      "proposer": {
			//        "address": "IhfskqTnO6eehukPpsQYBCnrfEE=",
			//        "pub_key": {
			//          "ed25519": "4Payp8jGB+/HF2opX32Ox+rwJ4O/WlCRz4sKwyuQfMA="
			//        },
			//        "voting_power": "6849070",
			//        "proposer_priority": "0"
			//      },
			//      "total_voting_power": "169513421"
			//    }
			//  },
			//  "signer": "atone1z7p3628vmw8psjwwzlcak0xqvgyssmwpa3guqm"
			//}
			chainID       = "beezee-1"
			timestamp     = toTime("2025-09-25T14:00:43.469479971Z")
			apphash       = b64Dec("XI9E+rmDbjwfn/EvQVp7CAqygaEI8Fingdh3xR+7aKw=")
			height        = uint64(19282029)
			blockhash     = b64Dec("q82PBaLn4Vg3oLm0v670BOm1Ay+CR69DyBqNA/V//l4=")
			parsethash    = b64Dec("PqBqLFCb6YhCQZFDnOeHzp1kHAbKs7fwHqOwnn8ha2s=")
			consensushash = b64Dec("BOEs4QnQgrOaBfgH2qUBI3M1r0rPLx++gYSRGuYVAns=")
			val1          = tendermint.NewValidator(
				"h9mY0OvyfYUX7rMfkXO+9bpEWdo=", "wF3Daj7tqX8nVOJd2WCkMVvqyx/lPzuI5/y3wwSRwbY=", 2,
			)
			val2 = tendermint.NewValidator(
				"mYLjIF9AxwR+BBlkD9M3f+xqKNE=", "dzCYrLu2sjpWSiEd2MVqsr+Q6ocBDUUUrRKBehGOeLM=", 3,
			)
			val3 = tendermint.NewValidator(
				"ysbGRHeOh4hD9fUbkM1RHP9bVfc=", "tYyou2DP3JWDrvKPtbAZpTCEtzqr9h1nsi/srNZwLiY=", 4,
			)
			valset        = tendermint.NewValset(val1, val2, val3)
			trustedValset = tendermint.NewValset(val1, val2, val3)
			signatures    = []tendermint.CommitSig{
				{
					BlockIDFlag:      tendermint.BlockIDFlagAbsent,
					ValidatorAddress: nil,
					Timestamp:        time.Time{},
					Signature:        nil,
				},
				{
					BlockIDFlag:      tendermint.BlockIDFlagCommit,
					ValidatorAddress: valset.Validators[1].Address,
					Timestamp:        toTime("2025-09-25T14:00:49.373544902Z"),
					Signature:        b64Dec("rgUFLD3Yk9JdcfWSqcscXeUZVY+Q7E1jQ1Eumkh/LT2fUI4dsk2NBeOhjyKaF5Ey/q6DwoINss6iSxNPj8s2Ag=="),
				},
				{
					BlockIDFlag:      tendermint.BlockIDFlagCommit,
					ValidatorAddress: valset.Validators[2].Address,
					Timestamp:        toTime("2025-09-25T14:00:49.396076476Z"),
					Signature:        b64Dec("thHpLNT6555JHoNr10iBpp8NJemRMPKMwnV4QDL6gLpaYChqrQND+8K4hQrC8Ld2IAuIT92ZHFWOmoVH9bIcBw=="),
				},
			}

			msg = tendermint.NewMsgHeader(
				chainID, timestamp, apphash, blockhash, parsethash, consensushash,
				height, types.Height{}, valset, trustedValset, signatures,
			)

			// expectedBytesToSign are computed from cometbft/types.Commit.VoteSignBytes() function
			expectedBytesToSign = []string{
				"220802116d382601000000002a0b088092b8c398feffffff0132086265657a65652d31",
				"6d0802116d3826010000000022480a20abcd8f05a2e7e15837a0b9b4bfaef404e9b5032f8247af43c81a8d03f57ffe5e1224080112203ea06a2c509be988424191439ce787ce9d641c06cab3b7f01ea3b09e7f216b6b2a0c089196d5c60610c6af8fb20132086265657a65652d31",
				"6d0802116d3826010000000022480a20abcd8f05a2e7e15837a0b9b4bfaef404e9b5032f8247af43c81a8d03f57ffe5e1224080112203ea06a2c509be988424191439ce787ce9d641c06cab3b7f01ea3b09e7f216b6b2a0c089196d5c60610bccbeebc0132086265657a65652d31",
			}
		)

		testVotesBytesToSign(t, msg, expectedBytesToSign)
	})
}

func testVotesBytesToSign(t *testing.T, msg *tendermint.MsgHeader, expectedBytesToSign []string) {
	for idx, val := range msg.ValidatorSet.Validators {
		bz := tendermint.VoteBytesToSign(
			msg.Header.ChainID, int64(msg.Commit.Height), int64(msg.Commit.Round),
			msg.Commit.BlockID.Hash, msg.Commit.BlockID.PartSetHeader.Total,
			msg.Commit.BlockID.PartSetHeader.Hash,
			msg.Commit.Signatures[idx].BlockIDFlag,
			msg.Commit.Signatures[idx].Timestamp,
		)
		if h := hex.EncodeToString(bz); h != expectedBytesToSign[idx] {
			t.Errorf("wrong bytes-to-sign, got %s, want %s", h, expectedBytesToSign[idx])
			continue
		}
		if msg.Commit.Signatures[idx].BlockIDFlag != tendermint.BlockIDFlagCommit {
			// skip non signed block
			continue
		}

		v := ed25519.Verify(val.PubKey, bz, msg.Commit.Signatures[idx].Signature)

		if !v {
			t.Errorf("failed to verify signature for validator index %d", idx)
		}
	}
}
