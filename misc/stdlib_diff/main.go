package main

import (
	"flag"
	"log"
)

func main() {
	var srcPath string
	var dstPath string
	var outDirectory string
	var srcIsGno bool

	flag.StringVar(&srcPath, "src", "", "Path to Gnoland standard libraries directory")
	flag.StringVar(&dstPath, "dst", "", "Path to Goland standard libraries directory")
	flag.StringVar(&outDirectory, "out", "", "Path to the directory where the report will be generated")
	flag.BoolVar(&srcIsGno, "src_is_gno", false, "If true, indicates that the src parameter correponds to the gno standard libraries")
	flag.Parse()

	reportBuilder, err := NewReportBuilder(srcPath, dstPath, outDirectory, srcIsGno)
	if err != nil {
		log.Fatal("can't build report builder: ", err.Error())
	}

	log.Println("Building report...")
	if err := reportBuilder.Build(); err != nil {
		log.Fatalln("can't build report: ", err.Error())
	}
	log.Println("Report generation done!")
}
