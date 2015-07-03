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
	data   [][]byte
	result [][]int32
}

var tests32 = []test32pair{
	{[][]byte{[]byte{0, 0, 0, 255, 255, 255, 128, 0, 0, 127, 255, 255}},
		[][]int32{[]int32{0, -1, -8388608, 8388607}},
	},
}

func TestToInt32(t *testing.T) {
	pair := tests32[0]
	res := toInt32(pair.data)
	for idx, val := range res {
		for idy, numval := range val {
			if numval != pair.result[idx][idy] {
				t.Error(
					"For", pair.data,
					"expected", pair.result,
					"got", res,
				)
			}
		}
	}
}

type test16pair struct {
	data   [][]byte
	result [][]int16
}

var tests16 = []test16pair{
	{[][]byte{[]byte{0x00, 0x00, 0xFF, 0xFF, 0x00, 0x80, 0xFF, 0x7F}},
		[][]int16{[]int16{0, -1, -32768, 32767}},
	},
}

func TestToInt16(t *testing.T) {
	pair := tests16[0]
	res, _ := toInt16(pair.data)
	for idx, val := range res {
		for idy, numval := range val {
			if numval != pair.result[idx][idy] {
				t.Error(
					"For", pair.data,
					"expected", pair.result,
					"got", res,
				)
			}
		}
	}
}

func TestToInt(t *testing.T) {
	pair16 := tests16[0]
	d16 := NewData(pair16.data)
	res, _ := d16.ToInt()
	for idx, val := range res {
		for idy, numval := range val {
			if numval != int(pair16.result[idx][idy]) {
				t.Error(
					"For", pair16.data,
					"expected", pair16.result,
					"got", res,
				)
			}
		}
	}

	pair32 := tests32[0]
	d32 := NewData(pair32.data)
	d32.SetNumBytes(3)
	res, _ = d32.ToInt()
	for idx, val := range res {
		for idy, numval := range val {
			if numval != int(pair32.result[idx][idy]) {
				t.Error(
					"For", pair32.data,
					"expected", pair32.result,
					"got", res,
				)
			}
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
	ns := 8
	nr := 256
	h, err := NewHeader(NumSignal("8"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	data := make([][]byte, ns)
	channel := make([]byte, ns*nr)
	for idx := range data {
		data[idx], channel = channel[:nr], channel[nr:]
	}
	for _, datum := range data {
		for idz, _ := range datum {
			datum[idz] = byte(idz)
		}
	}
	edf := NewEDF(h)
	edf.dataRecords = append(edf.dataRecords, NewData(data))
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
	nb := nbHeader + ns*nr
	if len(buf) != nb {
		t.Error("For TestWrite\n",
			"Expected: ", nb,
			"From header: ", edf.header.numbytes,
			"From data: ", ns*nr,
			"Got: ", len(buf))
	}
}

func TestUnmarshal(t *testing.T) {
	ns := 8
	nr := 256
	strArr := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	h, err := NewHeader(Version("0"), LocalPatientID("foo"),
		LocalRecordID("foo"), Startdate("foo"), Starttime("foo"), NumDataRecord("1"),
		Duration("foo"), NumSignal("8"), Reserved(""), Labels(strArr),
		TransducerTypes(strArr), PhysicalDimensions(strArr),
		PhysicalMins(strArr), PhysicalMaxs(strArr),
		DigitalMins(strArr), DigitalMaxs(strArr),
		Prefilters(strArr), NumSamples(strArr), NSReserved(strArr))

	if err != nil {
		t.Error("For TestUnmarshal %s\n", err)
		return
	}

	data := make([][]byte, ns)
	channel := make([]byte, ns*nr)
	for idx := range data {
		data[idx], channel = channel[:nr], channel[nr:]
	}
	for _, datum := range data {
		for idz, _ := range datum {
			datum[idz] = byte(idz)
		}
	}
	edf := NewEDF(h)
	edf.dataRecords = append(edf.dataRecords, NewData(data))
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
		for idx, sample := range data.signals {
			for idz, val := range sample {
				if val != edf.dataRecords[idy].signals[idx][idz] {
					t.Error("For TestWrite\n",
						"Expected: ", edf.dataRecords[idy].signals[idx][idz],
						"Got: ", val)
				}
			}
		}
	}
}
