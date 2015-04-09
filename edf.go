package edf

/*
HEADER RECORD (we suggest to also adopt the 12 simple additional EDF+ specs)
8 ascii : version of this data format {0: edf, 1: edf+, 2: bdf+?????}
80 ascii : local patient identification (mind item 3 of the additional EDF+ specs)
80 ascii : local recording identification (mind item 4 of the additional EDF+ specs)
8 ascii : startdate of recording (dd.mm.yy) (mind item 2 of the additional EDF+ specs)
8 ascii : starttime of recording (hh.mm.ss)
8 ascii : number of bytes in header record
44 ascii : reserved
8 ascii : number of data records (-1 if unknown, obey item 10 of the additional EDF+ specs)
8 ascii : duration of a data record, in seconds
4 ascii : number of signals (ns) in data record
ns * 16 ascii : ns * label (e.g. EEG Fpz-Cz or Body temp) (mind item 9 of the additional EDF+ specs)
ns * 80 ascii : ns * transducer type (e.g. AgAgCl electrode)
ns * 8 ascii : ns * physical dimension (e.g. uV or degreeC)
ns * 8 ascii : ns * physical minimum (e.g. -500 or 34)
ns * 8 ascii : ns * physical maximum (e.g. 500 or 40)
ns * 8 ascii : ns * digital minimum (e.g. -2048)
ns * 8 ascii : ns * digital maximum (e.g. 2047)
ns * 80 ascii : ns * prefiltering (e.g. HP:0.1Hz LP:75Hz)
ns * 8 ascii : ns * nr of samples in each data record
ns * 32 ascii : ns * reserved
DATA RECORD
nr of samples[1] * integer : first signal in the data record
nr of samples[2] * integer : second signal
..
..
nr of samples[ns] * integer : last signal
*/

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

var errNotPrintable = errors.New("outside the printable range")

// NewReader instantiates a new edf reader
func NewReader(reader io.Reader) io.Reader {
	return &Reader{
		reader: reader,
	}
}

// Reader holds edf data read from readable
type Reader struct {
	reader io.Reader
	header *Header
	data   *Data
}

// Read efd data from readable into Reader
func (r *Reader) Read(buf []byte) (n int, err error) {
	var ns int
  var innerIdx int
  var outterIdx int
  offset := r.header.GetOffsetMap()

	n, err = r.reader.Read(buf)
	if err != nil {
		return 0, err
	}

	for idx, val := range buf {
		switch {
		case idx < offset["version"]:
			r.header.version[idx] = val
		case idx < offset["LPID"]:
			r.header.LPID[idx-offset["version"]] = val // edf+ requires formatting
		case idx < offset["LRID"]:
			r.header.LRID[idx-offset["LPID"]] = val // edf+ requires formatting
		case idx < offset["startdate"]:
			r.header.startdate[idx-offset["LRID"]] = val // edf+ formatting
		case idx < offset["starttime"]:
			r.header.starttime[idx-offset["startdate"]] = val // edf+ 
		case idx < offset["numbytes"]:
			r.header.numbytes[idx-offset["starttime"]] = val // Check numbytes
		case idx < offset["reserved"]:
			r.header.reserved[idx-offset["numbytes"]] = val
		case idx < offset["numdatar"]:
			r.header.numdatar[idx-offset["reserved"]] = val
		case idx < offset["duration"]:
			r.header.duration[idx-offset["numdatar"]] = val
		case idx < offset["ns"]:
			r.header.ns[idx-offset["duration"]] = val

    // Done with the non-variable length part of header. Moving on
    // to header items that have a length that depends on ns.
		case idx == offset["ns"]:
			ns, err = asciiToInt(r.header.ns[:])
      // Allocate variable length header elements
      r.header.allocate(ns)
			if err != nil {
				return 0, err
			}
      offset["label"] = ns * len(r.header.label[0]) + offset["ns"]
			fallthrough
		case idx < offset["label"]:
      innerIdx = whichIndex(idx, offset["label"], len(r.header.label))
			r.header.label[innerIdx][idx-offset["ns"]] = val
    case idx == offset["label"]:
      offset["transducerType"] = ns * len(r.header.transducerType[0]) + offset["label"]
      fallthrough
    case idx < offset["transducerType"]:
      innerIdx = whichIndex(idx, offset["transducerType"], len(r.header.transducerType))
      r.header.transducerType[innerIdx][idx] = val
    case idx == offset["transducerType"]:
      offset["phydim"] = ns * len(r.header.phydim[0]) + offset["transducerType"]
      fallthrough
    case idx < offset["phydim"]:
      innerIdx = whichIndex(idx, offset["phydim"], len(r.header.phydim))
      r.header.phydim[innerIdx][0] = val
    case idx == offset["phydim"]:
      offset["phymin"] = ns * len(r.header.phymin[0]) + offset["phydim"]
    case idx < offset["phymin"]:
      innerIdx = whichIndex(idx, offset["phymin"], len(r.header.phymin))
      r.header.phymin[innerIdx] = val
    case idx == offset["phymin"]:
      offset["phymax"] = ns * len(r.header.phymax[0]) + offset["phymin"]
    case idx < offset["phymax"]:
      innerIdx = whichIndex(idx, offset["phymax"], len(r.header.phymax))
      r.header.phymax[innerIdx] = val
    case idx == offset["phymax"]:
      offset["digmin"] = ns * len(r.header.digmin[0]) + offset["digmin"]
    case idx < offset["digmin"]:
      innerIdx = whichIndex(idx, offset["digmin"], len(r.header.digmin))
      r.header.digmin[innerIdx] = val
    case idx == offset["digmin"]:
      offset["digmax"] = ns * len(r.header.digmax[0]) + offset["digmax"]
    case idx < offset["digmax"]:
      innerIdx = whichIndex(idx, offset["digmax"], len(r.header.digmax))
      r.header.digmax[innerIdx] = val
    case idx == offset["digmin"]:
      offset["prefilter"] = ns * len(r.header.prefilter[0]) + offset["prefilter"]
    case idx < offset["prefilter"]:
      innerIdx = whichIndex(idx, offset["prefilter"], len(r.header.prefilter))
      r.header.prefilter[innerIdx] = val
    case idx == offset["prefilter"]:
      offset["numsample"] = ns * len(r.header.numsample[0]) + offset["numsample"]
    case idx < offset["numsample"]:
      innerIdx = whichIndex(idx, offset["numsample"], len(r.header.numsample))
      r.header.numsample[innerIdx] = val
    case idx == offset["numsample"]:
      offset["nsreserved"] = ns * len(r.header.nsreserved[0]) + offset["nsreserved"]
    case idx < offset["nsreserved"]:
      innerIdx = whichIndex(idx, offset["nsreserved"], len(r.header.nsreserved)) 
      r.header.nsreserved[innerIdx] = val
    break
		}
	}
	return n, nil
}

func whichIndex(currIdx int, nsOffset int, arrayLen int) (arrIdx int) {
  return (currIdx - nsOffset) / arrayLen
}

func asciiToInt(ascii []byte) (n int, err error) {
	sArr := make([]string, len(ascii))
	for idx, val := range ascii {
		sArr[idx] = string(val)
	}
	s := strings.Join(sArr, "")
	n, err = strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// NewWriter instantiates a new edf writer
func NewWriter(w io.Writer, header *Header, data *Data) io.Writer {
	return &Writer{
		writer: w,
		header: header,
		data:   data,
	}
}

// Writer holds edf data send to writable
type Writer struct {
	writer io.Writer
	header *Header
	data   *Data
}

// Write edf data from Writer into writable
func (w *Writer) Write([]byte) (n int, err error) {
	headerContents, err := w.header.GetContents()
	if err != nil {
		return 0, err
	}
	n0, err := w.writer.Write(headerContents)
	if err != nil {
		return 0, err
	}
	dataContents := w.data.GetContents()
	n1, err := w.writer.Write(dataContents)
	if err != nil {
		return 0, err
	}
	n = n0 + n1
	return n, nil
}

// Version setter
func Version(number string) func(*Header) error {
	return func(h *Header) error {
		return h.setVersion(number)
	}
}

// LocalPatientID setter
func LocalPatientID(LPID string) func(*Header) error {
	return func(h *Header) error {
		return h.setLPID(LPID)
	}
}

// LocalRecordID setter
func LocalRecordID(LRID string) func(*Header) error {
	return func(h *Header) error {
		return h.setLRID(LRID)
	}
}

// Startdate setter
func Startdate(startdate string) func(*Header) error {
	return func(h *Header) error {
		return h.setStartdate(startdate)
	}
}

// Starttime setter
func Starttime(starttime string) func(*Header) error {
	return func(h *Header) error {
		return h.setStarttime(starttime)
	}
}

// Duration setter
func Duration(dur string) func(*Header) error {
	return func(h *Header) error {
		return h.setDuration(dur)
	}
}

// NumSignals setter for number of signals in data record
func NumSignals(ns string) func(*Header) error {
	return func(h *Header) error {
		return h.setNumSignals(ns)
	}
}

// Labels setter
func Labels(labels []string) func(*Header) error {
	return func(h *Header) error {
		return h.setLabels(labels)
	}
}

// TransducerTypes setter
func TransducerTypes(transducerTypes []string) func(*Header) error {
	return func(h *Header) error {
		return h.setTransducerTypes(transducerTypes)
	}
}

// PhysicalDimensions setter
func PhysicalDimensions(phydims []string) func(*Header) error {
	return func(h *Header) error {
		return h.setPhysicalDimensions(phydims)
	}
}

// PhysicalMins setter
func PhysicalMins(phymins []string) func(*Header) error {
	return func(h *Header) error {
		return h.setPhysicalMins(phymins)
	}
}

// PhysicalMaxs setter
func PhysicalMaxs(phymaxs []string) func(*Header) error {
	return func(h *Header) error {
		return h.setPhysicalMaxs(phymaxs)
	}
}

// DigitalMins setter
func DigitalMins(digmins []string) func(*Header) error {
	return func(h *Header) error {
		return h.setDigitalMins(digmins)
	}
}

// DigitalMaxs setter
func DigitalMaxs(digmaxs []string) func(*Header) error {
	return func(h *Header) error {
		return h.setDigitalMaxs(digmaxs)
	}
}

// Prefilters setter
func Prefilters(prefilters []string) func(*Header) error {
	return func(h *Header) error {
		return h.setPrefilters(prefilters)
	}
}

// NumSamples setter
func NumSamples(numsamples []string) func(*Header) error {
	return func(h *Header) error {
		return h.setNumSamples(numsamples)
	}
}

// NewHeader instantiates a edf header with optional funcational parameters to set values
// If optional parameters are left blank, we write 0 val bytes (ASCII NUL) into edf file
func NewHeader(numDataRecords int, options ...func(*Header) error) (*Header, error) {
	nr := strconv.Itoa(numDataRecords)
	numdatar := [8]byte{}
	for idx, val := range nr {
		if val < 32 || val > 126 {
			return nil, errNotPrintable
		}
		numdatar[idx] = nr[idx]
	}
	numbytes := numDataRecords*(16+80+8+8+8+8+8+80+16) + 8 + 80 + 80 + 8 + 8 + 8 + 44 + 8 + 8 + 4
	nb := strconv.Itoa(numbytes)
	numbytesArray := [8]byte{}
	for idx, val := range nb {
		if val < 32 || val > 126 {
			return nil, errNotPrintable
		}
		numbytesArray[idx] = nb[idx]
	}
	h := &Header{
		numdatar: numdatar,
		numbytes: numbytesArray,
	}
	for _, option := range options {
		if err := option(h); err != nil {
			return nil, err
		}
	}
	return h, nil
}

// Header holds edf header data by type
type Header struct {
	version        [8]byte
	LPID           [80]byte
	LRID           [80]byte
	startdate      [8]byte
	starttime      [8]byte
	numbytes       [8]byte
	reserved       [44]byte
	numdatar       [8]byte
	duration       [8]byte
	ns             [4]byte
	label          [][16]byte
	transducerType [][80]byte
	phydim         [][8]byte
	phymin         [][8]byte
	phymax         [][8]byte
	digmin         [][8]byte
	digmax         [][8]byte
	prefilter      [][80]byte
	numsample      [][8]byte
	nsreserved     [][16]byte
}

func (h *Header) setVersion(number string) error {
	for idx, val := range number {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.version[idx] = number[idx]
	}
	return nil
}

func (h *Header) setLPID(localPatientID string) error {
	for idx, val := range localPatientID {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.LPID[idx] = localPatientID[idx]
	}
	return nil
}

func (h *Header) setLRID(localRecordID string) error {
	for idx, val := range localRecordID {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.LRID[idx] = localRecordID[idx]
	}
	return nil
}

func (h *Header) setStartdate(startdate string) error {
	for idx, val := range startdate {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.startdate[idx] = startdate[idx]
	}
	return nil
}

func (h *Header) setStarttime(starttime string) error {
	for idx, val := range starttime {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.starttime[idx] = starttime[idx]
	}
	return nil
}

func (h *Header) setNumDataRecord(numdatar string) error {
	for idx, val := range numdatar {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.numdatar[idx] = numdatar[idx]
	}
	return nil
}

func (h *Header) setDuration(dur string) error {
	for idx, val := range dur {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.duration[idx] = dur[idx]
	}
	return nil
}

func (h *Header) setNumSignals(ns string) error {
	for idx, val := range ns {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.ns[idx] = ns[idx]
	}
	return nil
}

func (h *Header) setLabels(labels []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.label = make([][16]byte, ns)
	for idz, label := range labels {
		for idx, val := range label {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.label[idz][idx] = label[idx]
		}
	}
	return nil
}

func (h *Header) setTransducerTypes(tts []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.transducerType = make([][80]byte, ns)
	for idz, tt := range tts {
		for idx, val := range tt {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.transducerType[idz][idx] = tt[idx]
		}
	}
	return nil
}

func (h *Header) setPhysicalDimensions(phydims []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.phydim = make([][8]byte, ns)
	for idz, phydim := range phydims {
		for idx, val := range phydim {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.phydim[idz][idx] = phydim[idx]
		}
	}
	return nil
}

func (h *Header) setPhysicalMins(phymins []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.phymin = make([][8]byte, ns)
	for idz, phymin := range phymins {
		for idx, val := range phymin {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.phymin[idz][idx] = phymin[idx]
		}
	}
	return nil
}

func (h *Header) setPhysicalMaxs(phymaxs []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.phymax = make([][8]byte, ns)
	for idz, phymax := range phymaxs {
		for idx, val := range phymax {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.phymax[idz][idx] = phymax[idx]
		}
	}
	return nil
}

func (h *Header) setDigitalMins(digmins []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.digmin = make([][8]byte, ns)
	for idz, digmin := range digmins {
		for idx, val := range digmin {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.digmin[idz][idx] = digmin[idx]
		}
	}
	return nil
}

func (h *Header) setDigitalMaxs(digmaxs []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.digmax = make([][8]byte, ns)
	for idz, digmax := range digmaxs {
		for idx, val := range digmax {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.digmax[idz][idx] = digmax[idx]
		}
	}
	return nil
}

func (h *Header) setPrefilters(prefilters []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.prefilter = make([][80]byte, ns)
	for idz, prefilter := range prefilters {
		for idx, val := range prefilter {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.prefilter[idz][idx] = prefilter[idx]
		}
	}
	return nil
}

func (h *Header) setNumSamples(numsamples []string) error {
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return err
	}
	h.digmax = make([][8]byte, ns)
	for idz, numsample := range numsamples {
		for idx, val := range numsample {
			if val < 32 || val > 126 {
				return errNotPrintable
			}
			h.numsample[idz][idx] = numsample[idx]
		}
	}
	return nil
}

func (h *Header) GetOffsetMap() map[string]int {
  offset := make(map[string]int)
  offset["version"] = len(h.version)
  offset["LPID"] = len(h.LPID) + offset["version"]
  offset["LRID"] = len(h.LRID) + offset["LPID"]
  offset["startdate"] = len(h.startdate) + offset["LRID"]
  offset["starttime"] = len(h.starttime) + offset["startdate"]
  offset["numbytes"] = len(h.numbytes) + offset["starttime"]
  offset["reserved"] = len(h.reserved) + offset["numbytes"]
  offset["numdatar"] = len(h.numdatar) + offset["reserved"]
  offset["duration"] = len(h.duration) + offset["numdatar"]
  offset["ns"] = len(h.ns) + offset["duration"]
  return offset
}

func (h *Header) allocate(ns int) {
  h.label = make([][len(h.label[0])]byte, ns)
  h.transducerType = make([][len(h.transducerType[0])]byte, ns)
  h.phydim = make([][len(h.phydim[0])]byte, ns)
  h.phymin = make([][len(h.phymin[0])]byte, ns)
  h.phymax = make([][len(h.phymax[0])]byte, ns)
  h.digmin = make([][len(h.digmin[0])]byte, ns)
  h.digmax = make([][len(h.digmax[0])]byte, ns)
  h.prefilter = make([][len(h.prefilter[0])]byte, ns)
  h.numsample = make([][len(h.numsample[0])]byte, ns)
  h.nsreserved = make([][len(h.nsreserved[0])]byte, ns)
}

// GetContents of header in contiguous byte slice
func (h *Header) GetContents() (contents []byte, err error) {
	contents = append(contents, h.version[:]...)
	contents = append(contents, h.LPID[:]...)
	contents = append(contents, h.LRID[:]...)
	contents = append(contents, h.startdate[:]...)
	contents = append(contents, h.starttime[:]...)
	contents = append(contents, h.numbytes[:]...)
	contents = append(contents, h.reserved[:]...)
	contents = append(contents, h.numdatar[:]...)
	contents = append(contents, h.duration[:]...)
	contents = append(contents, h.ns[:]...)
	ns, err := strconv.Atoi(string(h.ns[:]))
	if err != nil {
		return contents, err
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.label[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.transducerType[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.phydim[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.phymin[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.phymax[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.digmin[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.digmax[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.prefilter[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.numsample[i][:]...)
	}
	for i := 0; i < ns; i++ {
		contents = append(contents, h.nsreserved[i][:]...)
	}
	return contents, nil
}

// NewData ...
func NewData(samples [][]byte) *Data {
	return &Data{
		samples: samples,
	}
}

// Data holds edf data record
type Data struct {
	samples [][]byte
}

// GetContents ...
func (d *Data) GetContents() (contents []byte) {
	var nilSep []byte
	contents = bytes.Join(d.samples, nilSep)
	return contents
}
