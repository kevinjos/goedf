package edf

import (
	"testing"
)

type test24pair struct {
	data   []byte
	result int32
}

var tests24 = []test24pair{
	{[]byte{0, 0, 0}, 0},
	{[]byte{255, 255, 255}, -1},
	{[]byte{128, 0, 0}, -8388608},
	{[]byte{127, 255, 255}, 8388607},
}

func TestConvert24bitTo32bit(t *testing.T) {
	for _, pair := range tests24 {
		res := convert24bitTo32bit(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type test32pair struct {
	data   []byte
	result []int32
}

var tests32 = []test32pair{
	{[]byte{0, 0, 0, 255, 255, 255, 128, 0, 0, 127, 255, 255}, []int32{0, -1, -8388608, 8388607}},
}

func TestToInt32(t *testing.T) {
	pair := tests32[0]
	res := toInt32(pair.data)
	for idx, val := range res {
		if val != pair.result[idx] {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type test16pair struct {
	data   []byte
	result []int16
}

var tests16 = []test16pair{
	{[]byte{0x00, 0x00, 0xFF, 0xFF, 0x00, 0x80, 0xFF, 0x7F}, []int16{0, -1, -32768, 32767}},
}

func TestToInt16(t *testing.T) {
	pair := tests16[0]
	res, _ := toInt16(pair.data)
	for idx, val := range res {
		if val != pair.result[idx] {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

func TestToInt(t *testing.T) {
	res, _ := ToInt(len(tests16[0].result), tests16[0].data)
	for idx, val := range res {
		if val != int(tests16[0].result[idx]) {
			t.Error(
				"For", tests16[0].data,
				"expected", tests16[0].result,
				"got", res,
			)
		}
	}

	res, _ = ToInt(len(tests16[0].result), tests32[0].data)
	for idx, val := range res {
		if val != int(tests32[0].result[idx]) {
			t.Error(
				"For", tests32[0].data,
				"expected", tests32[0].result,
				"got", res,
			)
		}
	}
}

type writer struct{}

func (w *writer) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type reader struct{}

func (r *reader) Read(buf []byte) (int, error) {
	return len(buf), nil
}

func TestMarshal(t *testing.T) {
	numsig, numsamp := 8, 256
	h, err := NewHeader(NumSignal("8"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	data := make([]byte, numsig*numsamp)
	for idx := range data {
		data[idx] = byte(idx % 256)
	}
	edf := NewEDF(h, []*Data{NewData(data)})

	buf, err := Marshal(edf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nbHeader, err := asciiToInt(edf.header.numbytes[:])
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nb := nbHeader + numsig*numsamp
	if len(buf) != nb {
		t.Error("For TestWrite\n",
			"Expected: ", nb,
			"From header: ", edf.header.numbytes,
			"From data: ", numsig*numsamp,
			"Got: ", len(buf))
	}
}

func TestUnmarshal(t *testing.T) {
	numsig := 8
	numsamp := 256
	strArr := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	strArrNsamp := []string{"256", "256", "256", "256", "256", "256", "256", "256"}
	h, err := NewHeader(Version("0"), LocalPatientID("foo"),
		LocalRecordID("foo"), Startdate("foo"), Starttime("foo"), NumDataRecord("1"),
		Duration("foo"), NumSignal("8"), Reserved(""), Labels(strArr),
		TransducerTypes(strArr), PhysicalDimensions(strArr),
		PhysicalMins(strArr), PhysicalMaxs(strArr),
		DigitalMins(strArr), DigitalMaxs(strArr),
		Prefilters(strArr), NumSamples(strArrNsamp), NSReserved(strArr))

	if err != nil {
		t.Error("For TestUnmarshal %s\n", err)
		return
	}

	data := make([]byte, numsig*numsamp)
	for idx := range data {
		data[idx] = byte(idx % 256)
	}
	edf := NewEDF(h, []*Data{NewData(data)})

	buf, err := Marshal(edf)
	if err != nil {
		t.Errorf("marshal in unmarshal: %s", err)
		return
	}

	newEDF, err := Unmarshal(buf)
	if err != nil {
		t.Errorf("unmarshal in unmarshal: %s", err)
		return
	}
	if newEDF.header.numbytes != edf.header.numbytes {
		t.Error("For TestWrite\n",
			"Expected: ", edf.header.numbytes,
			"Got: ", newEDF.header.numbytes)
	}
	if newEDF.header.numsignal != edf.header.numsignal {
		t.Error("For TestWrite\n",
			"Expected: ", edf.header.numsample,
			"Got: ", newEDF.header.numsample)
	}
	for idy, data := range newEDF.dataRecords {
		for idx, val := range data.rawData {
			if val != edf.dataRecords[idy].rawData[idx] {
				t.Error("For TestUnmarshal\n",
					"Expected: ", edf.dataRecords[idy].signals[idx],
					"Got: ", val)
			}
		}
	}
}
