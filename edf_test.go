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
	h, err := NewHeader(NS("8"))
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
	d := NewData(data)
	edf := NewEDF(h, d)
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
	genStrArr := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	h, err := NewHeader(Version("0"), LocalPatientID("foo"),
		LocalRecordID("foo"), Startdate("foo"), Starttime("foo"),
		Duration("foo"), NS("8"), Labels(genStrArr),
		TransducerTypes(genStrArr), PhysicalDimensions(genStrArr),
		PhysicalMins(genStrArr), PhysicalMaxs(genStrArr),
		DigitalMins(genStrArr), DigitalMaxs(genStrArr),
		Prefilters(genStrArr), NumSamples(genStrArr))
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
	d := NewData(data)
	edf := NewEDF(h, d)
	buf, err := Marshal(edf)

	//Unmarshal buf with new edf
	newH, err := NewHeader()
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	var newData [][]byte
	newD := NewData(newData)
	newEDF := NewEDF(newH, newD)
	err = Unmarshal(buf, newEDF)
	if newEDF.header.numbytes != edf.header.numbytes {
		t.Error("For TestWrite\n",
			"Expected: ", edf.header.numbytes,
			"Got: ", newEDF.header.numbytes)
	}
	if newEDF.header.ns != edf.header.ns {
		t.Error("For TestWrite\n",
			"Expected: ", edf.header.ns,
			"Got: ", newEDF.header.ns)
	}
	for idx, sample := range newEDF.data.samples {
		for idz, val := range sample {
			if val != edf.data.samples[idx][idz] {
				t.Error("For TestWrite\n",
					"Expected: ", edf.data.samples[idx][idz],
					"Got: ", val)
			}
		}
	}
}
