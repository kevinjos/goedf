package biosigio

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
	data   int32
	result []byte
}

var tests24rev = []test24revpair{
	{0, []byte{0, 0, 0}},
	{-1, []byte{255, 255, 255}},
	{-8388608, []byte{128, 0, 0}},
	{8388607, []byte{127, 255, 255}},
}

func TestConvert32IntTo3ByteArray(t *testing.T) {
	for _, pair := range tests24rev {
		res := convertInt32To3ByteArray(pair.data)
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
		a := convertInt32To3ByteArray(int32(i))
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

type writer struct{}

func (w *writer) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type reader struct{}

func (r *reader) Read(buf []byte) (int, error) {
	return len(buf), nil
}

func TestMarshalEDF(t *testing.T) {
	numsig, numsamp := 8, 256
	h, err := NewHeader(NumSignal("8"), NumDataRecord("1"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}

	signals := make([][]int16, numsig)
	for idx := range signals {
		signals[idx] = make([]int16, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = int16(rand.Intn(32767))
		}
	}

	edf := NewEDF(h, []*EDFData{&EDFData{Signals: signals}})

	buf, err := MarshalEDF(edf)
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

func TestMarshalBDF(t *testing.T) {
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

	signals := make([][]int32, numsig)
	for idx := range signals {
		signals[idx] = make([]int32, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = int32(rand.Intn(8388607))
		}
	}

	bdf := NewBDF(h, []*BDFData{&BDFData{Signals: signals}})
	buf, err := MarshalBDF(bdf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nbHeader, err := asciiToInt(bdf.Header.numbytes[:])
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	nb := nbHeader + numsig*numsamp*byteSize
	if len(buf) != nb {
		t.Error("For TestWrite\n",
			"Expected: ", nb,
			"From header: ", string(bdf.Header.numbytes[:]),
			"From data: ", numsig*numsamp*3,
			"Got: ", len(buf))
	}
}

func TestUnmarshalEDF(t *testing.T) {
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

	signals := make([][]int16, numsig)
	for idx := range signals {
		signals[idx] = make([]int16, numsamp)
		for idy := range signals[idx] {
			signals[idx][idy] = int16(rand.Intn(32767))
		}
	}

	edf := NewEDF(h, []*EDFData{&EDFData{Signals: signals}})
	buf, err := MarshalEDF(edf)
	if err != nil {
		t.Errorf("marshal in unmarshal: %s", err)
		return
	}

	newEDF, err := UnmarshalEDF(buf)
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

func TestUnmarshalEDFFile(t *testing.T) {
	fn := "./tstdata/testdata.edf"
	buf, err := ioutil.ReadFile(fn)
	if os.IsNotExist(err) {
		t.Logf("%s\nmissing EDF test data\n", err)
		return
	} else if err != nil {
		t.Errorf("read test data file: %s\n", err)
	}
	edf, err := UnmarshalEDF(buf)
	bufM, _ := MarshalEDF(edf)
	edf, err = UnmarshalEDF(bufM)
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
		drByteCount += numval * EDFDataByteSize
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

func TestUnmarshalBDFFile(t *testing.T) {
	fn := "./tstdata/testdata.bdf"
	buf, err := ioutil.ReadFile(fn)
	if os.IsNotExist(err) {
		t.Logf("%s\nmissing BDF test data\n", err)
		return
	} else if err != nil {
		t.Errorf("read test data file: %s\n", err)
	}
	bdf, err := UnmarshalBDF(buf)
	bufM, _ := MarshalBDF(bdf)
	bdf, err = UnmarshalBDF(bufM)
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
	headerByteCount, _ := asciiToInt(bdf.Header.numbytes[:])
	numDRs, _ := asciiToInt(bdf.Header.numdatar[:])
	var drByteCount int
	for _, val := range bdf.Header.numsample {
		numval, _ := asciiToInt(val[:])
		drByteCount += numval * BDFDataByteSize
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
