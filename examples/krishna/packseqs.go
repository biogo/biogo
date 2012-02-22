package main

import (
	"crypto/md5"
	"fmt"
	"github.com/kortschak/BioGo/align/pals"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
	"log"
	"os"
	"path/filepath"
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
		md5hash, _ := util.Hash(md5.New(), file)
		log.Printf("Reading %s: %s", fileName, fmt.Sprintf("%x", md5hash))

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
