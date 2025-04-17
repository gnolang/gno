package memberstore

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
)

var members MembersByTier
var Tiers TiersByName

const (
	T1 = "T1"
	T2 = "T2"
	T3 = "T3"
)

func init() {
	members = NewMembersByTier()

	Tiers = TiersByName{avl.NewTree()}
	Tiers.Set(T1, Tier{
		InvitationPoints: 3,
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 70
		},
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 3,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			return 3
		},
	})

	Tiers.Set(T2, Tier{
		InvitationPoints: 2,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) * 2
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return membersByTier.GetTierSize(T1) / 4
		},
		BasePower: 2,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t2ms := float64(membersByTier.GetTierSize(T2))
			t2, _ := tiersByName.GetTier(T2)

			t1p := t1.BasePower * t1ms
			t2p := t2.BasePower * t2ms

			// capped to 2/3 of tier 1
			t1ptreshold := t1p * (2.0 / 3.0)
			if t2p > t1ptreshold {
				return t1ptreshold / t2ms
			}

			return t2.BasePower
		},
	})

	Tiers.Set(T3, Tier{
		InvitationPoints: 1,
		MaxSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		MinSize: func(membersByTier MembersByTier, tiersByName TiersByName) int {
			return 0
		},
		BasePower: 1,
		PowerHandler: func(membersByTier MembersByTier, tiersByName TiersByName) float64 {
			t1ms := float64(membersByTier.GetTierSize(T1))
			t1, _ := tiersByName.GetTier(T1)
			t3ms := float64(membersByTier.GetTierSize(T3))
			t3, _ := tiersByName.GetTier(T3)

			t1p := t1.BasePower * t1ms
			t3p := t3.BasePower * t3ms

			//capped to 1/3 of tier 1
			t1ptreshold := t1p * (1.0 / 3.0)
			if t3p > t1ptreshold {
				return t1ptreshold / t3ms
			}

			return t3.BasePower
		},
	})

	members.SetTier(T1)
	members.SetMember(T1, std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), &Member{InvitationPoints: 3}) // Jae
	members.SetMember(T1, std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), &Member{InvitationPoints: 3}) // Manfred
	members.SetMember(T1, std.Address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), &Member{InvitationPoints: 3}) // Milos
	members.SetMember(T1, std.Address("g1qhskthp2uycmg4zsdc9squ2jds7yv3t0qyrlnp"), &Member{InvitationPoints: 3}) // Petar
	members.SetMember(T1, std.Address("g18amm3fc00t43dcxsys6udug0czyvqt9e7p23rd"), &Member{InvitationPoints: 3}) // Marc
	members.SetMember(T1, std.Address("g19p3yzr3cuhzqa02j0ce6kzvyjqfzwemw3vam0x"), &Member{InvitationPoints: 3}) // Guilhem
	members.SetMember(T1, std.Address("g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j"), &Member{InvitationPoints: 3}) // Ray
	members.SetMember(T1, std.Address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), &Member{InvitationPoints: 3}) // Maxwell
	members.SetMember(T1, std.Address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"), &Member{InvitationPoints: 3}) // Morgan
	members.SetMember(T1, std.Address("g1ker4vvggvsyatexxn3hkthp2hu80pkhrwmuczr"), &Member{InvitationPoints: 3}) // Sergio
	members.SetMember(T1, std.Address("g18x425qmujg99cfz3q97y4uep5pxjq3z8lmpt25"), &Member{InvitationPoints: 3}) // Antoine
	members.SetMember(T1, std.Address("g16tfrrul20g4jzt3z303raqw8vs8s2pqqh5clwu"), &Member{InvitationPoints: 3}) // Ilker
	members.SetMember(T1, std.Address("g1hy6zry03hg5d8le9s2w4fxme6236hkgd928dun"), &Member{InvitationPoints: 3}) // Jerónimo
	members.SetMember(T1, std.Address("g15ruzptpql4dpuyzej0wkt5rq6r26kw4nxu9fwd"), &Member{InvitationPoints: 3}) // Denis
	members.SetMember(T1, std.Address("g1lckl8j2g3jyyuq6fx7pke3uz4kemht7lw4fg5l"), &Member{InvitationPoints: 3}) // Danny
	members.SetMember(T1, std.Address("g1778y2yphxs2wpuaflsy5y9qwcd4gttn4g5yjx5"), &Member{InvitationPoints: 3}) // Michelle
	members.SetMember(T1, std.Address("g1mq7g0jszdmn4qdpc9tq94w0gyex37su892n80m"), &Member{InvitationPoints: 3}) // Alan
	members.SetMember(T1, std.Address("g197q5e9v00vuz256ly7fq7v3ekaun5cr7wmjgfh"), &Member{InvitationPoints: 3}) // Salvo
	members.SetMember(T1, std.Address("g1mpkp5lm8lwpm0pym4388836d009zfe4maxlqsq"), &Member{InvitationPoints: 3}) // Alexis
	members.SetMember(T1, std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), &Member{InvitationPoints: 3}) // Leon
	members.SetMember(T1, std.Address("g1whzkakk4hzjkvy60d5pwfk484xu67ar2cl62h2"), &Member{InvitationPoints: 3}) // Kirk
	members.SetMember(T1, std.Address("g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr"), &Member{InvitationPoints: 3}) // Albert

	members.SetTier(T2)
	members.SetMember(T2, std.Address("g1jazghxvvgz3egnr2fc8uf72z4g0l03596y9ls7"), &Member{InvitationPoints: 2}) // Nemanja
	members.SetMember(T2, std.Address("g1dfr24yhk5ztwtqn2a36m8f6ud8cx5hww4dkjfl"), &Member{InvitationPoints: 2}) // Antonio
	members.SetMember(T2, std.Address("g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay"), &Member{InvitationPoints: 2}) // Dongwon
	members.SetMember(T2, std.Address("g1r04aw56fgvzy859fachr8hzzhqkulkaemltr76"), &Member{InvitationPoints: 2}) // Blake
	members.SetMember(T2, std.Address("g17n4y745s08awwq4e0a38lagsgtntna0749tnxe"), &Member{InvitationPoints: 2}) // Jinwoo
	members.SetMember(T2, std.Address("g1ckae7tc5sez8ul3ssne75sk4muwgttp6ks2ky9"), &Member{InvitationPoints: 2}) // ByeongJun

	members.SetTier(T3)
	members.SetMember(T3, std.Address("g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a"), &Member{InvitationPoints: 1}) // Norman
	members.SetMember(T3, std.Address("g1qynsu9dwj9lq0m5fkje7jh6qy3md80ztqnshhm"), &Member{InvitationPoints: 1}) // Rémi
	members.SetMember(T3, std.Address("g17ernafy6ctpcz6uepfsq2js8x2vz0wladh5yc3"), &Member{InvitationPoints: 1}) // Dragos

}

func Render(string) string {
	var sb strings.Builder

	sb.WriteString("# Memberstore Govdao v3:\n\n")

	members.Iterate("", "", func(tn string, ti interface{}) bool {
		tree, ok := ti.(*avl.Tree)
		if !ok {
			return false
		}

		tier, ok := Tiers.GetTier(tn)
		if !ok {
			return false
		}

		tp := (tier.PowerHandler(members, Tiers) * float64(members.GetTierSize(tn)))

		sb.WriteString(ufmt.Sprintf("- Tier %v contains %v members with power: %v\n", tn, tree.Size(), tp))

		return false
	})

	return sb.String()
}

// Get gets the Members store
func Get() MembersByTier {
	realm := std.PreviousRealm().PkgPath()
	if !dao.InAllowedDAOs(realm) {
		panic("this Realm is not allowed to get the Members data: " + realm)
	}

	return members
}
