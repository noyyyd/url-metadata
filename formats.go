package urlmetadata

import (
	"fmt"
	"strings"
)

type videoFormat string

const (
	fMp4 videoFormat = "mp4"
	f3gp videoFormat = "3gp"
	fM4v videoFormat = "m4v"
	fMov videoFormat = "mov"
)

var fTypValuesForMP4 = []string{"avc1", "F4V", "F4P", "iso2", "isom", "mmp4", "mp41", "mp42", "NDSC", "NDSH", "NDSM", "NDSP", "NDSS", "NDXC", "NDXH", "NDXM", "NDXP", "NDXS"}
var fTypValuesFor3GP = []string{"3ge6", "3ge7", "3gg6", "3gp1", "3gp2", "3gp3", "3gp4", "3gp5", "3gp6", "3gs7"}
var fTypValuesForM4V = []string{"M4V", "M4VH", "M4VP"}
var fTypValuesForMOV = []string{"qt"}

func (vf videoFormat) fTypValues() []string {
	switch vf {
	case fMp4:
		return fTypValuesForMP4
	case f3gp:
		return fTypValuesFor3GP
	case fM4v:
		return fTypValuesForM4V
	case fMov:
		return fTypValuesForMOV
	default:
		return nil
	}
}

func (vf videoFormat) String() string {
	return string(vf)
}

func getFormat(rawFormat string) (videoFormat, error) {
	switch {
	case inSliceStr(fMp4.fTypValues(), rawFormat):
		return fMp4, nil
	case inSliceStr(f3gp.fTypValues(), rawFormat):
		return f3gp, nil
	case inSliceStr(fM4v.fTypValues(), rawFormat):
		return fM4v, nil
	case inSliceStr(fMov.fTypValues(), rawFormat):
		return fMov, nil
	default:
		return "", fmt.Errorf("unsupported video format: %s", rawFormat)
	}
}

func getFormatRegex() string {
	return fmt.Sprintf(
		"%s|%s|%s|%s",
		strings.Join(fMp4.fTypValues(), "|"),
		strings.Join(f3gp.fTypValues(), "|"),
		strings.Join(fM4v.fTypValues(), "|"),
		strings.Join(fMov.fTypValues(), "|"),
	)
}

func inSliceStr(slice []string, findVal string) bool {
	for _, val := range slice {
		if val == findVal {
			return true
		}
	}
	return false
}
