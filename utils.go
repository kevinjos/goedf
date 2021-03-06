package biosigio

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/kevinjos/eeg-web-server/int24"
)

func fillWithSpaces(input []byte) {
	for idx, _ := range input {
		input[idx] = '\x20'
	}
}

func asciiToInt(ascii []byte) (n int, err error) {
	sArr := make([]string, len(ascii))
	for idx, val := range ascii {
		if val == '\x00' || val == '\x20' {
			sArr = sArr[:idx]
			break
		}
		sArr[idx] = string(val)
	}
	s := strings.Join(sArr, "")
	n, err = strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func fixedHeaderOffsets() map[string]int {
	h, _ := NewHeader()
	offset := make(map[string]int)
	offset["version"] = len(h.version)
	offset["LPID"] = len(h.LPID)
	offset["LRID"] = len(h.LRID)
	offset["startdate"] = len(h.startdate)
	offset["starttime"] = len(h.starttime)
	offset["numbytes"] = len(h.numbytes)
	offset["reserved"] = len(h.reserved)
	offset["numdatar"] = len(h.numdatar)
	offset["duration"] = len(h.duration)
	offset["numsignal"] = len(h.numsignal)
	return offset
}

func variableHeaderOffsets(numsignal []byte) map[string]int {
	h, _ := NewHeader(NumSignal(string(numsignal)))
	ns, _ := asciiToInt(numsignal)
	offset := make(map[string]int)
	offset["label"] = len(h.label[0]) * ns
	offset["transducerType"] = len(h.transducerType[0]) * ns
	offset["phydim"] = len(h.phydim[0]) * ns
	offset["phymin"] = len(h.phymin[0]) * ns
	offset["phymax"] = len(h.phymax[0]) * ns
	offset["digmin"] = len(h.digmin[0]) * ns
	offset["digmax"] = len(h.digmax[0]) * ns
	offset["prefilter"] = len(h.prefilter[0]) * ns
	offset["numsample"] = len(h.numsample[0]) * ns
	offset["nsreserved"] = len(h.nsreserved[0]) * ns
	return offset
}

// toInt16 converts arrays of 2-byte two's complement little-endian integers
// to arrays of go int16
func toInt16(signal []byte) (res []int16, err error) {
	res = make([]int16, len(signal)/2)
	buf := bytes.NewReader(signal)
	if err = binary.Read(buf, binary.LittleEndian, res); err != nil {
		return res, err
	}
	return res, nil
}

// toInt32 converts arrays of 3-byte two's complement little-endian integers
// to arrays of go int32
func toInt32(signal []byte) (res []int32) {
	res = make([]int32, len(signal)/3)
	for idx, finished := 0, false; !finished; idx++ {
		if (idx+1)*3 == len(signal) {
			res[idx] = int24.UnmarshalSLE(signal[idx*3:])
			finished = true
		} else {
			res[idx] = int24.UnmarshalSLE(signal[idx*3 : (idx+1)*3])
		}
	}
	return res
}
