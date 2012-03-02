package main

import (
	"crypto/md5"
	"fmt"
	"github.com/kortschak/BioGo/align/pals"
	"github.com/kortschak/BioGo/io/seqio/fasta"
	"github.com/kortschak/BioGo/seq"
	"github.com/kortschak/BioGo/util"
	"os"
	"path/filepath"
)

func packSequence(fileName string) *seq.Seq {
	_, name := filepath.Split(fileName)
	packer := pals.NewPacker(name)

	if file, err := os.Open(fileName); err == nil {
		md5hash, _ := util.Hash(md5.New(), file)
		logger.Printf("Reading %s: %s", fileName, fmt.Sprintf("%x", md5hash))

		seqFile := fasta.NewReader(file)

		f, p := logger.Flags(), logger.Prefix()
		if verbose {
			logger.SetFlags(0)
			logger.SetPrefix("")
			logger.Println("Sequence            \t    Length\t   Bin Range")
		}

		var sequence *seq.Seq
		for {
			if sequence, err = seqFile.Read(); err == nil {
				if s := packer.Pack(sequence); verbose {
					logger.Println(s)
				}
			} else {
				break
			}
		}
		if verbose {
			logger.SetFlags(f)
			logger.SetPrefix(p)
		}
	} else {
		logger.Fatalf("Error: %v.\n", err)
	}

	packer.FinalisePack()

	return packer.Packed
}
