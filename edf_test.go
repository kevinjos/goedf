package edf

import (
	"io/ioutil"
	"math/rand"
	"os"
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

type test24revpair struct {
	data   int
	result []byte
}

var tests24rev = []test24revpair{
	{0, []byte{0, 0, 0}},
	{-1, []byte{255, 255, 255}},
	{-8388608, []byte{128, 0, 0}},
	{8388607, []byte{127, 255, 255}},
}

func TestConvertIntTo3ByteArray(t *testing.T) {
	for _, pair := range tests24rev {
		res := convertIntTo3ByteArray(pair.data)
		for idx, b := range res {
			if b != pair.result[idx] {
				t.Error(
					"For", pair.data,
					"expected", pair.result,
					"got", res,
				)
			}
		}
	}
}

func Test24bitToIntAndBack(t *testing.T) {
	for i := -8388608; i < 8388608; i++ {
		a := convertIntTo3ByteArray(i)
		b := convert24bitTo32bit(a)
		if i != int(b) {
			t.Error(
				"For", i,
				"Expected", i,
				"Got", b,
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
	res, _ := toInt(len(tests16[0].result), tests16[0].data)
	for idx, val := range res {
		if val != int(tests16[0].result[idx]) {
			t.Error(
				"For", tests16[0].data,
				"expected", tests16[0].result,
				"got", res,
			)
		}
	}

	res, _ = toInt(len(tests16[0].result), tests32[0].data)
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

func TestMarshal2ByteData(t *testing.T) {
	numsig, numsamp := 8, 256
	h, err := NewHeader(NumSignal("8"), NumDataRecord("1"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}

	signals := make([][]int, numsig)
	for idx := range signals {
		signals[idx] = make([]int, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = rand.Intn(32767)
		}
	}

	edf := NewEDF(h, []*Data{&Data{Signals: signals}})

	buf, err := Marshal(edf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nbHeader, err := asciiToInt(edf.Header.numbytes[:])
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nb := nbHeader + numsig*numsamp*2
	if len(buf) != nb {
		t.Error("For TestWrite\n",
			"Expected: ", nb,
			"From header: ", string(edf.Header.numbytes[:]),
			"From data: ", numsig*numsamp,
			"Got: ", len(buf))
	}
}

func TestMarshal3ByteData(t *testing.T) {
	numsig := 8
	numsamp := 256
	byteSize := 3
	strArr := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	strArrNsamp := []string{"256", "256", "256", "256", "256", "256", "256", "256"}
	strArrBSize := []string{"3", "3", "3", "3", "3", "3", "3", "3"}
	h, err := NewHeader(Version("0"), LocalPatientID("foo"),
		LocalRecordID("foo"), Startdate("foo"), Starttime("foo"), NumDataRecord("1"),
		Duration("foo"), NumSignal("8"), Reserved(""), Labels(strArr),
		TransducerTypes(strArr), PhysicalDimensions(strArr),
		PhysicalMins(strArr), PhysicalMaxs(strArr),
		DigitalMins(strArr), DigitalMaxs(strArr),
		Prefilters(strArr), NumSamples(strArrNsamp), NSReserved(strArrBSize))
	if err != nil {
		t.Error("For TestMarshal3ByteData\n", err)
		return
	}

	signals := make([][]int, numsig)
	for idx := range signals {
		signals[idx] = make([]int, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = rand.Intn(8388607)
		}
	}

	edf := NewEDF(h, []*Data{&Data{Signals: signals}})
	buf, err := Marshal(edf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nbHeader, err := asciiToInt(edf.Header.numbytes[:])
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nb := nbHeader + numsig*numsamp*byteSize
	if len(buf) != nb {
		t.Error("For TestWrite\n",
			"Expected: ", nb,
			"From header: ", string(edf.Header.numbytes[:]),
			"From data: ", numsig*numsamp*3,
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

	signals := make([][]int, numsig)
	for idx := range signals {
		signals[idx] = make([]int, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = rand.Intn(32767)
		}
	}

	edf := NewEDF(h, []*Data{&Data{Signals: signals}})
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
	if newEDF.Header.numbytes != edf.Header.numbytes {
		t.Error("For TestWrite\n",
			"Expected: ", edf.Header.numbytes,
			"Got: ", newEDF.Header.numbytes)
	}
	if newEDF.Header.numsignal != edf.Header.numsignal {
		t.Error("For TestWrite\n",
			"Expected: ", edf.Header.numsample,
			"Got: ", newEDF.Header.numsample)
	}
	for idy, data := range newEDF.DataRecords {
		for idx, val := range data.rawData {
			if val != edf.DataRecords[idy].rawData[idx] {
				t.Error("For TestUnmarshal\n",
					"Expected: ", edf.DataRecords[idy].rawData[idx],
					"Got: ", val)
			}
		}
	}
}

func TestUnmarshalFile(t *testing.T) {
	filenames := []string{"./tstdata/testdata.edf", "./tstdata/testdata3byte.edf"}
	for _, fn := range filenames {
		buf, err := ioutil.ReadFile(fn)
		if os.IsNotExist(err) {
			t.Logf("%s\nmissing 2-byte standard test data\n", err)
			continue
		} else if err != nil {
			t.Errorf("read test data file: %s\n", err)
		}
		edf, err := Unmarshal(buf)
		bytesize, err := asciiToInt(edf.Header.nsreserved[0][:])
		if err != nil {
			bytesize = 2
		}
		bufM, _ := Marshal(edf)
		edf, err = Unmarshal(bufM)
		if err != nil {
			t.Errorf("unmarshal test file: %s\n", err)
		}
		// Byte for byte comparison of file in to unmarshaled/marshaled buffer
		for idx, val := range buf {
			if val != bufM[idx] {
				t.Error("For TestUnmarshalFile\n",
					"Expected: ", val,
					"Got: ", bufM[idx],
					"At index: ", idx)
			}
		}
		// Check size of buffer against header vals
		bufMSize := len(bufM)
		bufSize := len(buf)
		headerByteCount, _ := asciiToInt(edf.Header.numbytes[:])
		numDRs, _ := asciiToInt(edf.Header.numdatar[:])
		var drByteCount int
		for _, val := range edf.Header.numsample {
			numval, _ := asciiToInt(val[:])
			drByteCount += numval * bytesize
		}
		if bufMSize != bufSize {
			t.Error("For TestUnmarshalFile\n",
				"Expected: ", bufSize,
				"Got: ", bufMSize)
		}
		byteCount := headerByteCount + drByteCount*numDRs
		if byteCount != bufSize {
			t.Error("For TestUnmarshalFile\n",
				"Expected: ", bufSize,
				"Got: ", drByteCount)
		}
	}
}
