package main

import (
	"os"
	"bio/seq"
	"bio/align/pals/dp"
	"bio/align/pals"
)

func WriteDPHits(w *pals.Writer, target, query *seq.Seq, hits []dp.DPHit, comp bool) (n int, err os.Error) {
	var pair *pals.FeaturePair

	for _, hit := range hits {
		if pair, err = pals.FeaturePairOf(target, query, hit, comp); err != nil {
			return
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
