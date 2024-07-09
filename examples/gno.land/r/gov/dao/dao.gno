package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	pproposal "gno.land/p/gov/proposal"
)

var (
	proposals = make([]*proposal, 0)
	members   = make([]std.Address, 0) // XXX: these should be pointers to avoid data duplication. Not possible due to VM bugs
)

const (
	msgMissingExecutor = "missing proposal executor"
	msgPropExecuted    = "prop already executed"
	msgPropExpired     = "prop is expired"
	msgPropInactive    = "prop is not active anymore"
	msgPropActive      = "prop is still active"
	msgPropNotAccepted = "prop is not accepted"

	msgCallerNotAMember = "caller is not member of govdao"
	msgProposalNotFound = "proposal not found"
)

func init() {
	// Prepare the initial members
	set := []std.Address{
		/* GNO CORE */
		"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", // Jae
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", // Manfred
		"g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2", // Milos
		"g1jazghxvvgz3egnr2fc8uf72z4g0l03596y9ls7", // Nemanja
		"g17p2kkyy9lp2z3ecw4psssk357vxp20afnyl00d", // Petar
		"g18amm3fc00t43dcxsys6udug0czyvqt9e7p23rd", // Marc
		"g1dfr24yhk5ztwtqn2a36m8f6ud8cx5hww4dkjfl", // Antonio
		"g19p3yzr3cuhzqa02j0ce6kzvyjqfzwemw3vam0x", // Guilhem
		"g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j", // Ray
		"g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq", // Maxwell
		"g1acn3xssksatydd0fcuslvgmjyw0fzkjdhusddg", // Dylan
		"g1cpx59z5r8vzeww2fm4ezpz7yvjs7kptywkm864", // Morgan
		"g1ker4vvggvsyatexxn3hkthp2hu80pkhrwmuczr", // Sergio
		"g125em6arxsnj49vx35f0n0z34putv5ty3376fg5", // Leon

		/* GNO DEVX */
		"g16tfrrul20g4jzt3z303raqw8vs8s2pqqh5clwu", // Ilker
		"g1hy6zry03hg5d8le9s2w4fxme6236hkgd928dun", // Jerónimo
		"g15ruzptpql4dpuyzej0wkt5rq6r26kw4nxu9fwd", // Denis
		"g1dnllrdzwfhxv3evyk09y48mgn5phfjvtyrlzm7", // Danny
		"g1778y2yphxs2wpuaflsy5y9qwcd4gttn4g5yjx5", // Michelle
		"g1mq7g0jszdmn4qdpc9tq94w0gyex37su892n80m", // Alan
		"g197q5e9v00vuz256ly7fq7v3ekaun5cr7wmjgfh", // Salvo
		"g1mpkp5lm8lwpm0pym4388836d009zfe4maxlqsq", // Alexis

		/* ONBLOC */
		"g17m8hlvm3k0agngz8vw29etpcjd2yvcel6pvt3k", // Andrew
		"g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay", // Dongwon
		"g1r04aw56fgvzy859fachr8hzzhqkulkaemltr76", // Blake
		"g17n4y745s08awwq4e0a38lagsgtntna0749tnxe", // Jinwoo
		"g1ckae7tc5sez8ul3ssne75sk4muwgttp6ks2ky9", // ByeongJun

		/* TERITORI */
		"g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a", // Norman

		/* BERTY */
		"g1qynsu9dwj9lq0m5fkje7jh6qy3md80ztqnshhm", // Rémi

		/* FLIPPANDO / ZENTASKTIC */
		"g17ernafy6ctpcz6uepfsq2js8x2vz0wladh5yc3", // Dragos
	}

	// Save the members
	members = append(members, set...)
}

type proposal struct {
	author   std.Address
	comment  string
	executor pproposal.Executor
	voter    Voter
	executed bool
	voters   []std.Address // XXX: these should be pointers to avoid data duplication. Not possible due to VM bugs.
}

func (p proposal) Status() Status {
	if p.executor.IsExpired() {
		return Expired
	}

	if p.executor.IsDone() {
		return Succeeded
	}

	if !p.voter.IsFinished(members) {
		return Active
	}

	if p.voter.IsAccepted(members) {
		return Accepted
	}

	return NotAccepted
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(comment string, executor pproposal.Executor) int {
	// XXX: require payment?
	if executor == nil {
		panic(msgMissingExecutor)
	}
	caller := std.GetOrigCaller() // XXX: CHANGE THIS WHEN MSGRUN PERSIST CODE ESCAPING THE main() SCOPE! IT IS UNSAFE!
	AssertIsMember(caller)

	prop := &proposal{
		comment:  comment,
		executor: executor,
		author:   caller,
		voter:    NewPercentageVoter(66), // at least 2/3 must say yes
	}

	proposals = append(proposals, prop)

	return len(proposals) - 1
}

func VoteOnProposal(idx int, option string) {
	assertProposalExists(idx)
	caller := std.GetOrigCaller() // XXX: CHANGE THIS WHEN MSGRUN PERSIST CODE ESCAPING THE main() SCOPE! IT IS UNSAFE!
	AssertIsMember(caller)

	prop := getProposal(idx)

	if prop.executed {
		panic(msgPropExecuted)
	}

	if prop.executor.IsExpired() {
		panic(msgPropExpired)
	}

	if prop.voter.IsFinished(members) {
		panic(msgPropInactive)
	}

	prop.voter.Vote(members, caller, option)
}

func ExecuteProposal(idx int) {
	assertProposalExists(idx)
	prop := getProposal(idx)

	if prop.executed {
		panic(msgPropExecuted)
	}

	if prop.executor.IsExpired() {
		panic(msgPropExpired)
	}

	if !prop.voter.IsFinished(members) {
		panic(msgPropActive)
	}

	if !prop.voter.IsAccepted(members) {
		panic(msgPropNotAccepted)
	}

	prop.executor.Execute()
	prop.voters = members
	prop.executed = true
}

func IsMember(addr std.Address) bool {
	if len(members) == 0 { // special case for initial execution
		return true
	}

	for _, v := range members {
		if v == addr {
			return true
		}
	}

	return false
}

func AssertIsMember(addr std.Address) {
	if !IsMember(addr) {
		panic(msgCallerNotAMember)
	}
}

func Render(path string) string {
	if path == "" {
		if len(proposals) == 0 {
			return "No proposals found :(" // corner case
		}

		output := ""
		for idx, prop := range proposals {
			output += ufmt.Sprintf("- [%d](/r/gov/dao:%d) - %s (**%s**)(by %s)\n", idx, idx, prop.comment, string(prop.Status()), prop.author)
		}

		return output
	}

	// else display the proposal
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404"
	}

	if !proposalExists(idx) {
		return "404"
	}
	prop := getProposal(idx)

	vs := members
	if prop.executed {
		vs = prop.voters
	}

	output := ""
	output += ufmt.Sprintf("# Prop #%d", idx)
	output += "\n\n"
	output += prop.comment
	output += "\n\n"
	output += ufmt.Sprintf("Status: %s", string(prop.Status()))
	output += "\n\n"
	output += ufmt.Sprintf("Voting status: %s", prop.voter.Status(vs))
	output += "\n\n"
	output += ufmt.Sprintf("Author: %s", string(prop.author))
	output += "\n\n"

	return output
}

func getProposal(idx int) *proposal {
	if idx > len(proposals)-1 {
		panic(msgProposalNotFound)
	}

	return proposals[idx]
}

func proposalExists(idx int) bool {
	return idx >= 0 && idx <= len(proposals)
}

func assertProposalExists(idx int) {
	if !proposalExists(idx) {
		panic("invalid proposal id")
	}
}
