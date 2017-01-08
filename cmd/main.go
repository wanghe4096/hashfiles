package main

import (
	"github.com/wanghe4096/HashFiles"
	"flag"
	"log"
	"os"
)

var (
	IgnoreDir  string
	Dir        string
	OutputFile string
)

func init() {

}

func main() {
	flag.StringVar(&Dir, "d", ".", "to scan directory")
	flag.StringVar(&IgnoreDir, "i", "", "ignore directory")
	flag.StringVar(&OutputFile, "o", "sha1.log", "output log file")
	flag.Parse()

	fd, err := os.OpenFile(OutputFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err.Error())
		return
	}
	HashFiles.SetOutput(fd)
	HashFiles.Run(Dir, IgnoreDir)
}
