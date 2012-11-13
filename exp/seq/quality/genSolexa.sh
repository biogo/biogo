cat < phred.go | \
	gofmt -r 'Phred -> Solexa' | \
	gofmt -r 'NewPhred -> NewSolexa' | \
	gofmt -r 'Qphred -> Qsolexa' | \
	gofmt -r 'Qphreds -> Qsolexas' | \
	gofmt -r 'DecodeToQphred -> DecodeToQsolexa' | \
	gofmt -r 'Ephred -> Esolexa' \
> solexa.go
