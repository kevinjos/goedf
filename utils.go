package edf

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

func fillWithSpaces(input []byte) {
	for idx, _ := range input {
		input[idx] = '\x20'
	}
}

func strToInt(str string) (n int, err error) {
	sArr := make([]string, len(str))
	for idx, val := range str {
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
			res[idx] = convert24bitTo32bit(signal[idx*3:])
			finished = true
		} else {
			res[idx] = convert24bitTo32bit(signal[idx*3 : (idx+1)*3])
		}
	}
	return res
}

//conver24bitTo32bit takes a byte slice of len 3
//and converts the 24bit 2's complement integer
//to the type int32 representation
func convert24bitTo32bit(c []byte) int32 {
	x := int((int(c[0]) << 16) | (int(c[1]) << 8) | int(c[2]))
	if (x & 8388608) > 0 {
		x |= 4278190080
	} else {
		x &= 16777215
	}
	return int32(x)
}

// ToInt coverts arrays of 2,3-byte two's complement little-endian integers
// to arrays of go ints
func ToInt(numsample int, sample []byte) (res []int, err error) {
	bytesize := len(sample) / numsample
	res = make([]int, numsample)
	switch bytesize {
	case 2:
		tmp, err := toInt16(sample)
		if err != nil {
			return res, err
		}
		for idx, numval := range tmp {
			res[idx] = int(numval)
		}
	case 3:
		tmp := toInt32(sample)
		for idx, numval := range tmp {
			res[idx] = int(numval)
		}
	}
	return res, nil
}
