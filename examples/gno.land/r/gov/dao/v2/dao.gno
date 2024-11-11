package govdao

import (
	"std"
	"strconv"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/ufmt"
)

var (
	d       *simpledao.SimpleDAO  // the current active DAO implementation
	members membstore.MemberStore // the member store
)

func init() {
	var (
		set = []membstore.Member{
			{
				Address:     std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), // Jae
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"), // Manfred
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), // Milos
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1jazghxvvgz3egnr2fc8uf72z4g0l03596y9ls7"), // Nemanja
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1qhskthp2uycmg4zsdc9squ2jds7yv3t0qyrlnp"), // Petar
				VotingPower: 20,
			},
			{
				Address:     std.Address("g18amm3fc00t43dcxsys6udug0czyvqt9e7p23rd"), // Marc
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1dfr24yhk5ztwtqn2a36m8f6ud8cx5hww4dkjfl"), // Antonio
				VotingPower: 20,
			},
			{
				Address:     std.Address("g19p3yzr3cuhzqa02j0ce6kzvyjqfzwemw3vam0x"), // Guilhem
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j"), // Ray
				VotingPower: 20,
			},
			{
				Address:     std.Address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), // Maxwell
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1cpx59z5r8vzeww2fm4ezpz7yvjs7kptywkm864"), // Morgan
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1ker4vvggvsyatexxn3hkthp2hu80pkhrwmuczr"), // Sergio
				VotingPower: 20,
			},
			{
				Address:     std.Address("g18x425qmujg99cfz3q97y4uep5pxjq3z8lmpt25"), // Antoine
				VotingPower: 20,
			},
			// GNO DEVX
			{
				Address:     std.Address("g16tfrrul20g4jzt3z303raqw8vs8s2pqqh5clwu"), // Ilker
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1hy6zry03hg5d8le9s2w4fxme6236hkgd928dun"), // Jerónimo
				VotingPower: 20,
			},
			{
				Address:     std.Address("g15ruzptpql4dpuyzej0wkt5rq6r26kw4nxu9fwd"), // Denis
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1dnllrdzwfhxv3evyk09y48mgn5phfjvtyrlzm7"), // Danny
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1778y2yphxs2wpuaflsy5y9qwcd4gttn4g5yjx5"), // Michelle
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1mq7g0jszdmn4qdpc9tq94w0gyex37su892n80m"), // Alan
				VotingPower: 20,
			},
			{
				Address:     std.Address("g197q5e9v00vuz256ly7fq7v3ekaun5cr7wmjgfh"), // Salvo
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1mpkp5lm8lwpm0pym4388836d009zfe4maxlqsq"), // Alexis
				VotingPower: 20,
			},
			{
				Address:     std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // Leon
				VotingPower: 20,
			},
			{
				Address:     std.Address("g1whzkakk4hzjkvy60d5pwfk484xu67ar2cl62h2"), // Kirk
				VotingPower: 20,
			},
			// AiB
			{
				Address:     std.Address("g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr"), // Albert
				VotingPower: 20,
			},
			// ONBLOC
			{
				Address:     std.Address("g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay"), // Dongwon
				VotingPower: 10,
			},
			{
				Address:     std.Address("g1r04aw56fgvzy859fachr8hzzhqkulkaemltr76"), // Blake
				VotingPower: 10,
			},
			{
				Address:     std.Address("g17n4y745s08awwq4e0a38lagsgtntna0749tnxe"), // Jinwoo
				VotingPower: 10,
			},
			{
				Address:     std.Address("g1ckae7tc5sez8ul3ssne75sk4muwgttp6ks2ky9"), // ByeongJun
				VotingPower: 10,
			},
			// TERITORI
			{
				Address:     std.Address("g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a"), // Norman
				VotingPower: 5,
			},
			// BERTY
			{
				Address:     std.Address("g1qynsu9dwj9lq0m5fkje7jh6qy3md80ztqnshhm"), // Rémi
				VotingPower: 5,
			},
			// FLIPPANDO / ZENTASKTIC
			{
				Address:     std.Address("g17ernafy6ctpcz6uepfsq2js8x2vz0wladh5yc3"), // Dragos
				VotingPower: 5,
			},
		}
	)

	// Set the member store
	members = membstore.NewMembStore(membstore.WithInitialMembers(set))

	// Set the DAO implementation
	d = simpledao.New(members)
}

// Propose is designed to be called by another contract or with
// `maketx run`, not by a `maketx call`.
func Propose(request dao.ProposalRequest) uint64 {
	idx, err := d.Propose(request)
	if err != nil {
		panic(err)
	}

	return idx
}

// VoteOnProposal casts a vote for the given proposal
func VoteOnProposal(id uint64, option dao.VoteOption) {
	if err := d.VoteOnProposal(id, option); err != nil {
		panic(err)
	}
}

// ExecuteProposal executes the proposal
func ExecuteProposal(id uint64) {
	if err := d.ExecuteProposal(id); err != nil {
		panic(err)
	}
}

// GetPropStore returns the active proposal store
func GetPropStore() dao.PropStore {
	return d
}

// GetMembStore returns the active member store
func GetMembStore() membstore.MemberStore {
	return members
}

func Render(path string) string {
	if path == "" {
		numProposals := d.Size()

		if numProposals == 0 {
			return "No proposals found :(" // corner case
		}

		output := ""

		offset := uint64(0)
		if numProposals >= 10 {
			offset = uint64(numProposals) - 10
		}

		// Fetch the last 10 proposals
		for idx, prop := range d.Proposals(offset, uint64(10)) {
			output += ufmt.Sprintf(
				"- [Proposal #%d](%s:%d) - (**%s**)(by %s)\n",
				idx,
				"/r/gov/dao/v2",
				idx,
				prop.Status().String(),
				prop.Author().String(),
			)
		}

		return output
	}

	// Display the detailed proposal
	idx, err := strconv.Atoi(path)
	if err != nil {
		return "404: Invalid proposal ID"
	}

	// Fetch the proposal
	prop, err := d.ProposalByID(uint64(idx))
	if err != nil {
		return ufmt.Sprintf("unable to fetch proposal, %s", err.Error())
	}

	// Render the proposal
	output := ""
	output += ufmt.Sprintf("# Prop #%d", idx)
	output += "\n\n"
	output += prop.Render()
	output += "\n\n"

	return output
}
