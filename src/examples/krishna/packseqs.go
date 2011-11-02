package main

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"bio/seq"
	"bio/util"
	"bio/io/seqio/fasta"
	"bio/align/pals"
)

// temporary filler - write a logger
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Flags() int
	Output(calldepth int, s string) error
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Prefix() string
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	SetFlags(flag int)
	SetPrefix(prefix string)
	Verbose(verbose bool)
	IsVerbose() bool
	Debug(verbose bool)
	IsDebug() bool
}

func packSequence(fileName string) *seq.Seq {
	_, name := filepath.Split(fileName)
	packer := pals.NewPacker(name)

	if file, err := os.Open(fileName); err == nil {
		md5, _ := util.Hash(file)
		log.Printf("Reading %s: %s", fileName, fmt.Sprintf("%x", md5))

		seqFile := fasta.NewReader(file)

		f, p := log.Flags(), log.Prefix()
		if verbose {
			log.SetFlags(0)
			log.SetPrefix("")
			log.Println("Sequence            \t    Length\t   Bin Range")
		}

		var sequence *seq.Seq
		for {
			if sequence, err = seqFile.Read(); err == nil {
				if s := packer.Pack(sequence); verbose {
					log.Println(s)
				}
			} else {
				break
			}
		}
		if verbose {
			log.SetFlags(f)
			log.SetPrefix(p)
		}
	} else {
		if debug {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		}
		log.Fatalf("Error: %v.\n", err)
	}

	packer.FinalisePack()

	return packer.Packed
}
