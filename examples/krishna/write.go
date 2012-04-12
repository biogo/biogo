package main

import (
	"github.com/kortschak/biogo/align/pals"
	"github.com/kortschak/biogo/align/pals/dp"
	"github.com/kortschak/biogo/seq"
	"sync"
)

var wlock = &sync.Mutex{}

func WriteDPHits(w *pals.Writer, target, query *seq.Seq, hits []dp.DPHit, comp bool) (n int, err error) {
	wlock.Lock()
	defer wlock.Unlock()

	for _, hit := range hits {
		pair, err := pals.NewFeaturePair(target, query, hit, comp)
		if err != nil {
			return n, err
		} else {
			ln, err := w.Write(pair)
			n += ln
			if err != nil {
				return n, err
			}
		}
	}

	return
}
