package urlmetadata

import (
	"encoding/hex"
	"strconv"
)

const (
	ftyp = "ftyp" // ftyp - атом в котором хранится формат видео
	moov = "moov" // moov - атом в котором хранятся все метаданные видео
	mvhd = "mvhd" // mvhd - атом в котором хранятся метаданные о времени создания/изменения и длинне видео
)

type atom struct {
	version        byte
	offset         int64
	name           string
	rawBytesWalker bytesWalker
}

func (a *atom) setNextOffset(b []byte) {
	offset, _ := strconv.ParseInt(hex.EncodeToString(b[a.offset:a.offset+4]), 16, 64)

	a.offset += offset
}

func (a *atom) isSaveOffset(b []byte) bool {
	// вычитаем 4 байта чтобы убедиться что мы сможем получить имя атома
	return a.offset < int64(len(b)-4)
}

func (a *atom) setName(b []byte) {
	a.name = string(b[a.offset+4 : a.offset+8])
}

func (a *atom) isFTYP() bool {
	return a.name == ftyp
}

func (a *atom) isMOOV() bool {
	return a.name == moov
}

func (a *atom) isMVHD() bool {
	return a.name == mvhd
}

func (a *atom) skipBlockBytes(b []byte) {
	a.offset += 4
	a.setName(b)
}
