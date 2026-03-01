package loader

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

// this is only executed when loaded into genesis
func init() {
	memberstore.Get().SetTier(memberstore.T1)
	memberstore.Get().SetMember(memberstore.T1, address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), memberstore.NewMember(3)) // Jae
	memberstore.Get().SetMember(memberstore.T1, address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), memberstore.NewMember(3)) // Manfred
	memberstore.Get().SetMember(memberstore.T1, address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), memberstore.NewMember(3)) // Milos
	memberstore.Get().SetMember(memberstore.T1, address("g18amm3fc00t43dcxsys6udug0czyvqt9e7p23rd"), memberstore.NewMember(3)) // Marc
	memberstore.Get().SetMember(memberstore.T1, address("g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j"), memberstore.NewMember(3)) // Ray
	memberstore.Get().SetMember(memberstore.T1, address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"), memberstore.NewMember(3)) // Morgan
	memberstore.Get().SetMember(memberstore.T1, address("g1ker4vvggvsyatexxn3hkthp2hu80pkhrwmuczr"), memberstore.NewMember(3)) // Sergio
	memberstore.Get().SetMember(memberstore.T1, address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"), memberstore.NewMember(3)) // Antoine
	memberstore.Get().SetMember(memberstore.T1, address("g1hy6zry03hg5d8le9s2w4fxme6236hkgd928dun"), memberstore.NewMember(3)) // Jerónimo
	memberstore.Get().SetMember(memberstore.T1, address("g1lckl8j2g3jyyuq6fx7pke3uz4kemht7lw4fg5l"), memberstore.NewMember(3)) // Danny
	memberstore.Get().SetMember(memberstore.T1, address("g1778y2yphxs2wpuaflsy5y9qwcd4gttn4g5yjx5"), memberstore.NewMember(3)) // Michelle
	memberstore.Get().SetMember(memberstore.T1, address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), memberstore.NewMember(3)) // Leon

	memberstore.Get().SetTier(memberstore.T2)
	memberstore.Get().SetMember(memberstore.T2, address("g1jazghxvvgz3egnr2fc8uf72z4g0l03596y9ls7"), memberstore.NewMember(2)) // Nemanja
	memberstore.Get().SetMember(memberstore.T2, address("g1dfr24yhk5ztwtqn2a36m8f6ud8cx5hww4dkjfl"), memberstore.NewMember(2)) // Antonio
	memberstore.Get().SetMember(memberstore.T2, address("g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay"), memberstore.NewMember(2)) // Dongwon
	memberstore.Get().SetMember(memberstore.T2, address("g17n4y745s08awwq4e0a38lagsgtntna0749tnxe"), memberstore.NewMember(2)) // Jinwoo
	memberstore.Get().SetMember(memberstore.T2, address("g1mpkp5lm8lwpm0pym4388836d009zfe4maxlqsq"), memberstore.NewMember(2)) // Alexis
	memberstore.Get().SetMember(memberstore.T2, address("g197q5e9v00vuz256ly7fq7v3ekaun5cr7wmjgfh"), memberstore.NewMember(2)) // Salvo

	memberstore.Get().SetTier(memberstore.T3)
	memberstore.Get().SetMember(memberstore.T3, address("g1qynsu9dwj9lq0m5fkje7jh6qy3md80ztqnshhm"), memberstore.NewMember(1)) // Rémi
	memberstore.Get().SetMember(memberstore.T3, address("g1mq7g0jszdmn4qdpc9tq94w0gyex37su892n80m"), memberstore.NewMember(1)) // Alan
	memberstore.Get().SetMember(memberstore.T3, address("g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a"), memberstore.NewMember(1)) // Norman
	memberstore.Get().SetMember(memberstore.T3, address("g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr"), memberstore.NewMember(1)) // Albert
	memberstore.Get().SetMember(memberstore.T3, address("g1ckae7tc5sez8ul3ssne75sk4muwgttp6ks2ky9"), memberstore.NewMember(1)) // ByeongJun
	memberstore.Get().SetMember(memberstore.T3, address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), memberstore.NewMember(1)) // Maxwell
	memberstore.Get().SetMember(memberstore.T3, address("g19p3yzr3cuhzqa02j0ce6kzvyjqfzwemw3vam0x"), memberstore.NewMember(1)) // Guilhem

	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.GetInstance(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}
