// PKGPATH: gno.land/r/zrealm_import_std_in
// A realm imports sort: sorting its own slice through sort.Sort works.
package zrealm_import_std_in

import "sort"

type List []int

func (l List) Len() int           { return len(l) }
func (l List) Less(i, j int) bool { return l[i] < l[j] }
func (l List) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

var gList = List{3, 1, 2}

func SortAll(cur realm) {
	var s sort.Interface = gList
	sort.Sort(s)
}

func main(cur realm) {
	SortAll(cross(cur))
	println("sorted:", gList[0], gList[1], gList[2])
}

// Output:
// sorted: 1 2 3
