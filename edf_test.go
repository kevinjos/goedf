package edf

import (
	"testing"
)

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
