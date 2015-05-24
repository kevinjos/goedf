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
	"strconv"
	"strings"
)

var errNotPrintable = errors.New("outside the printable range")

// Unmarshal byteslice into edf
func Unmarshal(data []byte, edf *EDF) error {
	var ns int
	var outterIdx int
	offset := edf.header.GetOffsetMap()

	for idx, val := range data {
		switch {
		case idx < offset["version"]:
			edf.header.version[idx] = val
		case idx < offset["LPID"]:
			edf.header.LPID[idx-offset["version"]] = val // edf+ requires formatting
		case idx < offset["LRID"]:
			edf.header.LRID[idx-offset["LPID"]] = val // edf+ requires formatting
		case idx < offset["startdate"]:
			edf.header.startdate[idx-offset["LRID"]] = val // edf+ formatting
		case idx < offset["starttime"]:
			edf.header.starttime[idx-offset["startdate"]] = val // edf+
		case idx < offset["numbytes"]:
			edf.header.numbytes[idx-offset["starttime"]] = val // Check numbytes
		case idx < offset["reserved"]:
			edf.header.reserved[idx-offset["numbytes"]] = val
		case idx < offset["numdatar"]:
			edf.header.numdatar[idx-offset["reserved"]] = val
		case idx < offset["duration"]:
			edf.header.duration[idx-offset["numdatar"]] = val
		case idx < offset["ns"]:
			edf.header.ns[idx-offset["duration"]] = val

		// Done with the non-variable length part of header. Moving on
		// to header items that have a length that depends on ns.
		case idx == offset["ns"]:
			var err error
			ns, err = asciiToInt(edf.header.ns[:])
			// Allocate variable length header elements
			edf.header.allocate(ns)
			if err != nil {
				return err
			}
			offset["label"] = ns*len(edf.header.label[0]) + offset["ns"]
			fallthrough
		case idx < offset["label"]:
			la := len(edf.header.label[0])
			outterIdx = whichIndex(idx, offset["ns"], la)
			edf.header.label[outterIdx][idx%la] = val
		case idx == offset["label"]:
			offset["transducerType"] = ns*len(edf.header.transducerType[0]) + offset["label"]
			fallthrough
		case idx < offset["transducerType"]:
			la := len(edf.header.transducerType[0])
			outterIdx = whichIndex(idx, offset["label"], la)
			edf.header.transducerType[outterIdx][idx%la] = val
		case idx == offset["transducerType"]:
			offset["phydim"] = ns*len(edf.header.phydim[0]) + offset["transducerType"]
			fallthrough
		case idx < offset["phydim"]:
			la := len(edf.header.phydim[0])
			outterIdx = whichIndex(idx, offset["transducerType"], la)
			edf.header.phydim[outterIdx][idx%la] = val
		case idx == offset["phydim"]:
			offset["phymin"] = ns*len(edf.header.phymin[0]) + offset["phydim"]
			fallthrough
		case idx < offset["phymin"]:
			la := len(edf.header.phymin[0])
			outterIdx = whichIndex(idx, offset["phydim"], la)
			edf.header.phymin[outterIdx][idx%la] = val
		case idx == offset["phymin"]:
			offset["phymax"] = ns*len(edf.header.phymax[0]) + offset["phymin"]
			fallthrough
		case idx < offset["phymax"]:
			la := len(edf.header.phymax[0])
			outterIdx = whichIndex(idx, offset["phymin"], la)
			edf.header.phymax[outterIdx][idx%la] = val
		case idx == offset["phymax"]:
			offset["digmin"] = ns*len(edf.header.digmin[0]) + offset["phymax"]
			fallthrough
		case idx < offset["digmin"]:
			la := len(edf.header.digmin[0])
			outterIdx = whichIndex(idx, offset["phymax"], la)
			edf.header.digmin[outterIdx][idx%la] = val
		case idx == offset["digmin"]:
			offset["digmax"] = ns*len(edf.header.digmax[0]) + offset["digmin"]
			fallthrough
		case idx < offset["digmax"]:
			la := len(edf.header.digmax[0])
			outterIdx = whichIndex(idx, offset["digmin"], la)
			edf.header.digmax[outterIdx][idx%la] = val
		case idx == offset["digmax"]:
			offset["prefilter"] = ns*len(edf.header.prefilter[0]) + offset["digmax"]
			fallthrough
		case idx < offset["prefilter"]:
			la := len(edf.header.prefilter[0])
			outterIdx = whichIndex(idx, offset["digmax"], la)
			edf.header.prefilter[outterIdx][idx%la] = val
		case idx == offset["prefilter"]:
			offset["numsample"] = ns*len(edf.header.numsample[0]) + offset["prefilter"]
			fallthrough
		case idx < offset["numsample"]:
			la := len(edf.header.numsample[0])
			outterIdx = whichIndex(idx, offset["prefilter"], la)
			edf.header.numsample[outterIdx][idx%la] = val
		case idx == offset["numsample"]:
			offset["nsreserved"] = ns*len(edf.header.nsreserved[0]) + offset["numsample"]
			fallthrough
		case idx < offset["nsreserved"]:
			la := len(edf.header.nsreserved[0])
			outterIdx = whichIndex(idx, offset["numsample"], la)
			edf.header.nsreserved[outterIdx][idx%la] = val
			break
		}
	}
	return edf
}

// Marshal edf into byte slice
func Marshal(edf *EDF) ([]byte, error) {
	var data []byte
	data, err := edf.header.AppendContents(data)
	data = edf.data.AppendContents(data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func whichIndex(currIdx int, nsOffset int, arrayLen int) (arrIdx int) {
	return (currIdx - nsOffset) / arrayLen
}

func asciiToInt(ascii []byte) (n int, err error) {
	sArr := make([]string, len(ascii))
	for idx, val := range ascii {
		if val == '\x00' {
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

// Duration setter
func NS(ns string) func(*Header) error {
	return func(h *Header) error {
		return h.setNS(ns)
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

// NewHeader instantiates a edf header
// Optional funcational parameters to set values in header
// If optional parameters are left blank, default is ASCII NUL
func NewHeader(options ...func(*Header) error) (*Header, error) {
	h := &Header{}
	for _, option := range options {
		if err := option(h); err != nil {
			return nil, err
		}
	}
	// The caller should specify the sample number
	// The reader will not know yet
	// The writer will, and should pass SetNS() to NewHeader
	if h.ns == [4]byte{} {
		return h, nil
	}
	ns, err := asciiToInt(h.ns[:])
	if err != nil {
		return nil, err
	}
	h.allocate(ns)
	nb := h.calcNumBytes(ns)
	h.setNumBytes(strconv.Itoa(nb))
	return h, nil
}

func (h *Header) calcNumBytes(ns int) (nb int) {
	nb += len(h.version)
	nb += len(h.LPID)
	nb += len(h.LRID)
	nb += len(h.startdate)
	nb += len(h.starttime)
	nb += len(h.numbytes)
	nb += len(h.reserved)
	nb += len(h.numdatar)
	nb += len(h.duration)
	nb += len(h.ns)
	nb += ns * len(h.label[0])
	nb += ns * len(h.transducerType[0])
	nb += ns * len(h.phydim[0])
	nb += ns * len(h.phymin[0])
	nb += ns * len(h.phymax[0])
	nb += ns * len(h.digmin[0])
	nb += ns * len(h.digmax[0])
	nb += ns * len(h.prefilter[0])
	nb += ns * len(h.numsample[0])
	nb += ns * len(h.nsreserved[0])
	return nb
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

func (h *Header) setNumBytes(numbytes string) error {
	for idx, val := range numbytes {
		if val < 32 || val > 126 {
			return errNotPrintable
		}
		h.numbytes[idx] = numbytes[idx]
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

func (h *Header) setNS(ns string) error {
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
	h.numsample = make([][8]byte, ns)
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

// AppendContents of header in contiguous byte slice
func (h *Header) AppendContents(contents []byte) (buf []byte, err error) {
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
	ns, err := asciiToInt(h.ns[:])
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
	buf = contents
	return buf, nil
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
func (d *Data) AppendContents(buf []byte) []byte {
	var nilSep []byte
	buf = append(buf, bytes.Join(d.samples, nilSep)...)
	return buf
}

func NewEDF(h *Header, d *Data) *EDF {
	return &EDF{
		header: h,
		data:   d,
	}
}

type EDF struct {
	header *Header
	data   *Data
}
