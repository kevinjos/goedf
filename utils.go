package edf

import (
	"strconv"
	"strings"
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
