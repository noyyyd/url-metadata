package urlmetadata

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	// reqSize - судя по тестам делать замер запроса больше не имеет смысла,
	// так как для длинных видео(10+ минут) чаще всего мы можем найти нужный
	// атом за один запрос, а короткие в большинстве своём требуют ещё один запрос (бывает по разному конечно)
	reqSize = 128

	// reqCount - количество попыток повторения запроса, в случае его подвисания.
	reqCount = 3
)

const (
	// offsetSize - размер смещения
	offsetSize = 4
	// nameSize - размер имени(типа) атома
	nameSize = 4
	// mvhdMaxSize - максимально возможный размер данных атома mvhd c нужными нам данными
	mvhdMaxSize = 40

	// payloadSize - максимально возможный размер атома mvhd + смещение и имя родительского атома moov
	payloadSize = offsetSize + nameSize + mvhdMaxSize
)

// Metadata описывает получаемую из файла информацию
type Metadata struct {
	Format       videoFormat
	DateCreate   time.Time
	DateModified time.Time
	Duration     int
	FileSize     int64
}

// GetMetadataVideo - функция для получения метаданных видео по ссылке
func GetMetadataVideo(cli http.Client, videoUrl string) (metadata *Metadata, err error) {
	var start int64
	var end int64 = reqSize

	b, fileSize, err := doRangeBytesWithLength(cli, videoUrl, start, end)
	if err != nil {
		return nil, err
	}

	format, err := getVideoFormat(b)
	if err != nil {
		return nil, err
	}

	for {
		mvhd, needSmallReq := findMVHD(b)
		if mvhd.isMVHD() {
			parsedMVHD, err := parseMVHD(mvhd)
			if err != nil {
				return nil, err
			}

			metadata = new(Metadata)

			metadata.Format = format
			metadata.FileSize = fileSize
			metadata.DateCreate = parsedMVHD.DateCreate
			metadata.DateModified = parsedMVHD.DateModified
			metadata.Duration = int(math.Ceil(float64(parsedMVHD.Duration) / float64(parsedMVHD.Time)))

			return metadata, nil
		} else {
			start = mvhd.offset
			end = mvhd.offset + reqSize

			if needSmallReq {
				end = mvhd.offset + payloadSize
			}
		}

		b, err = doRangeBytes(cli, videoUrl, start, end)
		if err != nil {
			return nil, err
		}
	}
}

// getVideoFormat - получает формат видео из самого первого атома ftyp
func getVideoFormat(b []byte) (format videoFormat, err error) {
	a := new(atom)
	a.setName(b)

	if a.isFTYP() {
		a.setNextOffset(b)
		a.rawBytesWalker.setRaw(b[offsetSize+nameSize : a.offset])
	}

	r, err := regexp.Compile(getFormatRegex())
	if err != nil {
		return "", err
	}

	rawFormat := r.FindString(string(a.rawBytesWalker.getRaw()))

	return getFormat(rawFormat)
}

// findMVHD - перебирает переданный набор байт в поиске атома с имене moov, который хранит в себе метаданные видео.
// В случае если в переданном наборе байт не был найден moov, возвращает последний найденный атом с его смещением.
func findMVHD(b []byte) (lastAtom *atom, needSmallReq bool) {
	lastAtom = new(atom)

	for {
		// сначала берём имя атома, так как имя мы берём с зависимости от значения offset
		lastAtom.setName(b)

		if lastAtom.isMOOV() {
			// в теории может возникнуть ситуация когда мы скачали килобайт, а атом moov находится
			// в самом конце и обрезан в нужном нам месте. Тут детектим этот кейс и
			// и если байт не хватает, делаем ещё один запрос на максимально возможный размер mvhd
			if len(b[lastAtom.offset:]) < payloadSize {
				return lastAtom, true
			}

			// пропускаем смещение атома moov и его имя, так как они нас не интересуют.
			// mvhd должен быть первым вложенным атомом moov, поэтому пропустим только 2 блока байт
			for !lastAtom.isMVHD() {
				lastAtom.skipBlockBytes(b)
			}

			// для удобства сохраняем байты с начала контента атома mvhd
			lastAtom.rawBytesWalker.setRaw(b[lastAtom.offset:])

			return lastAtom, false
		}

		// offset изменяем в последнюю очередь
		lastAtom.setNextOffset(b)

		if !lastAtom.isSaveOffset(b) {
			return lastAtom, false
		}
	}
}

func doRangeBytes(cli http.Client, videoUrl string, start, end int64) (b []byte, err error) {
	var resp *http.Response

	for i := 0; i < reqCount; i++ {
		resp, err = doRange(cli, videoUrl, start, end)
		if err != nil {
			continue
		}

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	return b, nil
}

func doRangeBytesWithLength(cli http.Client, videoUrl string, start, end int64) (b []byte, fileSize int64, err error) {
	var resp *http.Response

	for i := 0; i < reqCount; i++ {
		resp, err = doRange(cli, videoUrl, start, end)
		if err != nil {
			continue
		}

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		return nil, 0, err
	}

	resp, err = doHead(cli, videoUrl)
	if err != nil {
		return nil, 0, err
	}

	contentLength := resp.Header.Get("Content-Length")

	fileSize, err = strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return nil, 0, err
	}

	return b, fileSize, nil
}

func doRange(cli http.Client, videoUrl string, start, end int64) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, videoUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err = cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func doHead(cli http.Client, videoUrl string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodHead, videoUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err = cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
