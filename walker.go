package urlmetadata

type bytesWalker struct {
	index int
	raw   []byte
}

func (bw *bytesWalker) nextRaw(i int) []byte {
	defer func() {
		bw.index += i
	}()
	return bw.raw[bw.index : bw.index+i]
}

func (bw *bytesWalker) getRaw() []byte {
	return bw.raw
}

func (bw *bytesWalker) setRaw(b []byte) {
	bw.raw = b
}
