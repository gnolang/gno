package main

import (
	"flag"
	"log"
)

func main() {
	var srcPath string
	var dstPath string
	var outDirectory string

	flag.StringVar(&srcPath, "src", "", "Path to Gnoland standard libraries directory")
	flag.StringVar(&dstPath, "dst", "", "Path to Goland standard libraries directory")
	flag.StringVar(&outDirectory, "out", "", "Path to the directory where the report will be generated")
	flag.Parse()

	reportBuilder, err := NewReportBuilder(srcPath, dstPath, outDirectory)
	if err != nil {
		log.Fatal("can't build report builder: ", err.Error())
	}

	log.Println("Building report...")
	if err := reportBuilder.Build(); err != nil {
		log.Fatalln("can't build report: ", err.Error())
	}
	log.Println("Report generation done!")
}
