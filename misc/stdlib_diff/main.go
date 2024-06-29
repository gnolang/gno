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

	flag.StringVar(&srcPath, "src", "", "Directory containing packages that will be compared to destination")
	flag.StringVar(&dstPath, "dst", "", "Directory containing packages; used to compare src packages")
	flag.StringVar(&outDirectory, "out", "", "Directory where the report will be created")
	flag.BoolVar(&srcIsGno, "src_is_gno", false, "If true, indicates that the src parameter corresponds to the gno standard libraries")
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
