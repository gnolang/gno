package client

type BaseOptions struct {
	Home                  string
	Remote                string
	Quiet                 bool
	InsecurePasswordStdin bool
	Config                string
}

var DefaultBaseOptions = BaseOptions{
	Home:                  "",
	Remote:                "127.0.0.1:26657",
	Quiet:                 false,
	InsecurePasswordStdin: false,
	Config:                "",
}
