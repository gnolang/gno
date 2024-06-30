package consensus

import (
	"bufio"
	goerrors "errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/events"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	// event bus subscriber
	subscriber = "replay-file"
)

// --------------------------------------------------------
// replay messages interactively or all at once

// Replay msgs in file or start the console
func (cs *ConsensusState) ReplayFile(file string, console bool) error {
	if cs.IsRunning() {
		return errors.New("cs is already running, cannot replay")
	}
	if cs.wal != nil {
		return errors.New("cs wal is open, cannot replay")
	}

	cs.startForReplay()

	// ensure all new step events are regenerated as expected

	newStepSub := events.SubscribeToEvent(cs.evsw, subscriber, cstypes.EventNewRoundStep{})
	defer cs.evsw.RemoveListener(subscriber)

	// just open the file for reading, no need to use wal
	fp, err := os.OpenFile(file, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}

	pb := newPlayback(file, fp, cs, cs.state.Copy())
	defer pb.fp.Close() //nolint: errcheck

	var nextN int // apply N msgs in a row
	var msg *walm.TimedWALMessage
	var meta *walm.MetaMessage
	for {
		if nextN == 0 && console {
			nextN = pb.replayConsoleLoop()
		}

		msg, meta, err = pb.dec.ReadMessage()
		if goerrors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}

		if err := pb.cs.readReplayMessage(msg, meta, newStepSub); err != nil {
			return err
		}

		if nextN > 0 {
			nextN--
		}
		pb.count++
	}
}

// ------------------------------------------------
// playback manager

type playback struct {
	cs *ConsensusState

	fp    *os.File
	dec   *walm.WALReader
	count int // how many lines/msgs into the file are we

	// replays can be reset to beginning
	fileName     string   // so we can close/reopen the file
	genesisState sm.State // so the replay session knows where to restart from
}

func newPlayback(fileName string, fp *os.File, cs *ConsensusState, genState sm.State) *playback {
	return &playback{
		cs:           cs,
		fp:           fp,
		fileName:     fileName,
		genesisState: genState,
		dec:          walm.NewWALReader(fp, maxMsgSize),
	}
}

// go back count steps by resetting the state and running (pb.count - count) steps
func (pb *playback) replayReset(count int, newStepSub <-chan events.Event) error {
	pb.cs.Stop()
	pb.cs.Wait()

	newCS := NewConsensusState(pb.cs.config, pb.genesisState.Copy(), pb.cs.blockExec,
		pb.cs.blockStore, pb.cs.txNotifier)
	newCS.SetEventSwitch(pb.cs.evsw)
	newCS.startForReplay()

	if err := pb.fp.Close(); err != nil {
		return err
	}
	fp, err := os.OpenFile(pb.fileName, os.O_RDONLY, 0o600)
	if err != nil {
		return err
	}
	pb.fp = fp
	pb.dec = walm.NewWALReader(fp, maxMsgSize)
	count = pb.count - count
	fmt.Printf("Reseting from %d to %d\n", pb.count, count)
	pb.count = 0
	pb.cs = newCS
	var msg *walm.TimedWALMessage
	var meta *walm.MetaMessage
	for i := 0; i < count; i++ {
		msg, meta, err = pb.dec.ReadMessage()
		if goerrors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
		if err := pb.cs.readReplayMessage(msg, meta, newStepSub); err != nil {
			return err
		}
		pb.count++
	}
	return nil
}

func (cs *ConsensusState) startForReplay() {
	cs.Logger.Error("Replay commands are disabled until someone updates them and writes tests")
	/* TODO:!
	// since we replay tocks we just ignore ticks
		go func() {
			for {
				select {
				case <-cs.tickChan:
				case <-cs.Quit:
					return
				}
			}
		}()*/
}

// console function for parsing input and running commands
func (pb *playback) replayConsoleLoop() int {
	for {
		fmt.Printf("> ")
		bufReader := bufio.NewReader(os.Stdin)
		line, more, err := bufReader.ReadLine()
		if more {
			osm.Exit("input is too long")
		} else if err != nil {
			osm.Exit(err.Error())
		}

		tokens := strings.Split(string(line), " ")
		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "next":
			// "next" -> replay next message
			// "next N" -> replay next N messages

			if len(tokens) == 1 {
				return 0
			}
			i, err := strconv.Atoi(tokens[1])
			if err != nil {
				fmt.Println("next takes an integer argument")
			} else {
				return i
			}

		case "back":
			// "back" -> go back one message
			// "back N" -> go back N messages

			// NOTE: "back" is not supported in the state machine design,
			// so we restart and replay up to

			// ensure all new step events are regenerated as expected

			newStepSub := events.SubscribeToEvent(pb.cs.evsw, subscriber, cstypes.EventNewRoundStep{})
			defer pb.cs.evsw.RemoveListener(subscriber)

			if len(tokens) == 1 {
				if err := pb.replayReset(1, newStepSub); err != nil {
					pb.cs.Logger.Error("Replay reset error", "err", err)
				}
			} else {
				i, err := strconv.Atoi(tokens[1])
				if err != nil {
					fmt.Println("back takes an integer argument")
				} else if i > pb.count {
					fmt.Printf("argument to back must not be larger than the current count (%d)\n", pb.count)
				} else if err := pb.replayReset(i, newStepSub); err != nil {
					pb.cs.Logger.Error("Replay reset error", "err", err)
				}
			}

		case "rs":
			// "rs" -> print entire round state
			// "rs short" -> print height/round/step
			// "rs <field>" -> print another field of the round state

			rs := pb.cs.RoundState
			if len(tokens) == 1 {
				fmt.Println(rs)
			} else {
				switch tokens[1] {
				case "short":
					fmt.Printf("%v/%v/%v\n", rs.Height, rs.Round, rs.Step)
				case "validators":
					fmt.Println(rs.Validators)
				case "proposal":
					fmt.Println(rs.Proposal)
				case "proposal_block":
					fmt.Printf("%v %v\n", rs.ProposalBlockParts.StringShort(), rs.ProposalBlock.StringShort())
				case "locked_round":
					fmt.Println(rs.LockedRound)
				case "locked_block":
					fmt.Printf("%v %v\n", rs.LockedBlockParts.StringShort(), rs.LockedBlock.StringShort())
				case "votes":
					fmt.Println(rs.Votes.StringIndented("  "))

				default:
					fmt.Println("Unknown option", tokens[1])
				}
			}
		case "n":
			fmt.Println(pb.count)
		}
	}
}
