package edf

import (
	"fmt"
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

func TestWrite(t *testing.T) {
	h, err := NewHeader(NS("8"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	data := make([][]byte, 8)
	channel := make([]byte, 8*256)
	for idx := range data {
		data[idx], channel = channel[:256], channel[256:]
	}
	for _, datum := range data {
		for idz, _ := range datum {
			datum[idz] = byte(idz)
		}
	}
	d := NewData(data)
	w := NewWriter(&writer{}, h, d)
	var buf []byte
	n, err := w.Write(buf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	if len(buf) != n {
		t.Error("For TestWrite\n",
			"Expected: ", n,
			"Got: ", len(buf))
	}
}

func TestWriteRead(t *testing.T) {
	h, err := NewHeader(NS("8"))
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	data := make([][]byte, 8)
	channel := make([]byte, 8*256)
	for idx := range data {
		data[idx], channel = channel[:256], channel[256:]
	}
	for _, datum := range data {
		for idz, _ := range datum {
			datum[idz] = byte(idz)
		}
	}
	d := NewData(data)
	w := NewWriter(&writer{}, h, d)
	var buf []byte
	_, err = w.Write(buf)
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}

	hReader, err := NewHeader()
	if err != nil {
		t.Error("For TestWrite\n", err)
		return
	}
	dataReader := make([][]byte, 8)
	dReader := NewData(dataReader)
	r := NewReader(&reader{}, hReader, dReader)
	_, err = r.Read(buf)
	if err != nil {
		t.Error("For TestWriteRead:", err)
		return
	}
	fmt.Printf("Header: %+v\n", hReader)
	if hReader.numbytes != h.numbytes {
		t.Error("For TestWriteRead\n",
			"Expected: ", h.numbytes,
			"Got: ", hReader.numbytes)
	}
}
