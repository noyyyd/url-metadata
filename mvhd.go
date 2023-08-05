package urlmetadata

import (
	"encoding/binary"
	"fmt"
	"time"
)

type mvhdAtom struct {
	Offset       []byte
	Name         string
	DateCreate   time.Time
	DateModified time.Time
	Time         uint32
	Duration     uint64

	IsBigStep bool
}

func parseMVHD(a *atom) (*mvhdAtom, error) {
	if !a.isMVHD() {
		return nil, fmt.Errorf("an atom is not mvhd, an atom is %s", a.name)
	}

	var step int

	parsedMVHD := new(mvhdAtom)

	parsedMVHD.Offset = a.rawBytesWalker.nextRaw(4)
	parsedMVHD.Name = string(a.rawBytesWalker.nextRaw(4))
	parsedMVHD.IsBigStep = parsedMVHD.isBigStep(a.rawBytesWalker.nextRaw(4))

	if parsedMVHD.IsBigStep {
		step = 8
	} else {
		step = 4
	}

	parsedMVHD.DateCreate = parsedMVHD.getDate(a.rawBytesWalker.nextRaw(step))
	parsedMVHD.DateModified = parsedMVHD.getDate(a.rawBytesWalker.nextRaw(step))
	parsedMVHD.Time = binary.BigEndian.Uint32(a.rawBytesWalker.nextRaw(4))
	parsedMVHD.Duration = parsedMVHD.uint(a.rawBytesWalker.nextRaw(step))

	return parsedMVHD, nil
}

func (ma *mvhdAtom) isBigStep(versionBytes []byte) bool {
	return versionBytes[len(versionBytes)-1] == 1
}

var macUTCepoch = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)

func (ma *mvhdAtom) getDate(dateBytes []byte) time.Time {
	return macUTCepoch.Add(time.Duration(ma.uint(dateBytes)) * time.Second)
}

func (ma *mvhdAtom) uint(dateBytes []byte) uint64 {
	if ma.IsBigStep {
		return binary.BigEndian.Uint64(dateBytes)
	} else {
		return uint64(binary.BigEndian.Uint32(dateBytes))
	}
}
