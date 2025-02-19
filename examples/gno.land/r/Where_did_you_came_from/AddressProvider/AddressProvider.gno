package AddressProvider

import (
	"std"
	"ufmt"

	"gno.land/p/demo/avl"
)

// SCW 정보를 저장할 구조체
type SCWInfo struct {
	owner      std.Address // 실제 사용자 주소
	scwAddress std.Address // SCW 주소
	createdAt  uint64      // 생성 시점의 블록 높이
}

// 상태 변수들
var (
	// user address => SCW address 매핑
	userToSCW = avl.NewTree()
	// SCW address => SCW 정보 매핑
	scwDetails = avl.NewTree()
	// 생성된 총 SCW 수
	totalSCWs uint64 = 0
)

// SCW가 있는지 확인하고 없으면 새로 생성하는 함수
func GetOrCreateSCW() std.Address {
	caller := std.GetOrigCaller()

	// 기존 SCW가 있는지 확인
	if scwAddr, exists := GetSCW(caller); exists {
		return scwAddr
	}

	// 새 SCW 생성
	return createNewSCW(caller)
}

// 특정 사용자의 SCW 주소 조회
func GetSCW(userAddr std.Address) (std.Address, bool) {
	value, exists := userToSCW.Get(userAddr.String())
	if !exists {
		return std.Address(""), false
	}
	return std.Address(value.(string)), true
}

// 새로운 SCW 생성 (내부 함수)
func createNewSCW(owner std.Address) std.Address {
	// SCW 주소 생성 (예: "g1_scw_" + totalSCWs를 문자열로 변환)
	scwAddr := std.Address(fmt.Sprintf("g1_scw_%d", totalSCWs))

	// SCW 정보 생성
	scwInfo := SCWInfo{
		owner:      owner,
		scwAddress: scwAddr,
	}

	// 매핑 정보 저장
	userToSCW.Set(owner.String(), scwAddr.String())
	scwDetails.Set(scwAddr.String(), scwInfo)

	// 카운터 증가
	totalSCWs++

	return scwAddr
}

// SCW 상세 정보 조회
func GetSCWInfo(scwAddr std.Address) (SCWInfo, bool) {
	value, exists := scwDetails.Get(scwAddr.String())
	if !exists {
		return SCWInfo{}, false
	}
	return value.(SCWInfo), true
}

// 총 생성된 SCW 수 조회
func GetTotalSCWs() uint64 {
	return totalSCWs
}

// 특정 SCW의 소유자 확인
func IsOwner(userAddr std.Address, scwAddr std.Address) bool {
	info, exists := GetSCWInfo(scwAddr)
	if !exists {
		return false
	}
	return info.owner == userAddr
}
