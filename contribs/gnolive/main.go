package main

func main() {
	scriptFile := ".script.gno"
	go inotify(scriptFile, onUpdate)
	os.Exec("$EDITOR", scriptFile)
}

func onUpdate() {
	body := parseFile()
	a := ast.Parse(body)
	goImports(a)
	// remove comments
	src := ast.Generate(a)
	os.Exec("gnokey maketx run -file=", src)
}
