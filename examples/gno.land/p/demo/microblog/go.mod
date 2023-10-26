module github.com/gnolang/gno/examples/gno.land/p/demo/microblog

require (
	github.com/gnolang/gno/examples/gno.land/p/demo/avl v0.0.0-latest
	github.com/gnolang/gno/examples/gno.land/p/demo/ufmt v0.0.0-latest
	github.com/gnolang/gno/examples/gno.land/r/demo/users v0.0.0-latest
)

replace gno.land/p/demo/avl v0.0.0-latest => /home/howl/.config/gno/pkg/mod/gno.land/p/demo/avl

replace gno.land/p/demo/ufmt v0.0.0-latest => /home/howl/.config/gno/pkg/mod/gno.land/p/demo/ufmt

replace gno.land/r/demo/users v0.0.0-latest => /home/howl/.config/gno/pkg/mod/gno.land/r/demo/users
