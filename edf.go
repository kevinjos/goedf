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
	"errors"
	"fmt"
	"strconv"
)

var errNotPrintable = errors.New("outside the printable range")

// Unmarshal byteslice into edf
func Unmarshal(buf []byte) (edf *EDF, err error) {
	foffset := fixedHeaderOffsets()
	fixed := func(os int) (x, res []byte) {
		return buf[:os], buf[os:]
	}

	version, buf := fixed(foffset["version"])
	lpid, buf := fixed(foffset["LPID"])
	lrid, buf := fixed(foffset["LRID"])
	startdate, buf := fixed(foffset["startdate"])
	starttime, buf := fixed(foffset["starttime"])
	numbytes, buf := fixed(foffset["numbytes"])
	reserved, buf := fixed(foffset["reserved"])
	numdatar, buf := fixed(foffset["numdatar"])
	duration, buf := fixed(foffset["duration"])
	numsignal, buf := fixed(foffset["numsignal"])

	ns, err := asciiToInt(numsignal)
	if err != nil {
		fmt.Errorf("serialize ascii to int failure %v, for %v\n", err, numsignal)
	}

	voffset := variableHeaderOffsets(numsignal)
	variable := func(os int) (x []string, res []byte) {
		x = make([]string, ns)
		osps := os / ns
		for idx, _ := range x {
			x[idx] = string(buf[:osps])
			buf = buf[osps:]
		}
		return x, buf
	}

	label, buf := variable(voffset["label"])
	transducerType, buf := variable(voffset["transducerType"])
	phydim, buf := variable(voffset["phydim"])
	phymin, buf := variable(voffset["phymin"])
	phymax, buf := variable(voffset["phymax"])
	digmin, buf := variable(voffset["digmin"])
	digmax, buf := variable(voffset["digmax"])
	prefilter, buf := variable(voffset["prefilter"])
	numsample, buf := variable(voffset["numsample"])
	nsreserved, buf := variable(voffset["nsreserved"])

	h, err := NewHeader(Version(string(version)),
		LocalPatientID(string(lpid)),
		LocalRecordID(string(lrid)),
		Startdate(string(startdate)),
		Starttime(string(starttime)),
		NumBytes(string(numbytes)),
		Reserved(string(reserved)),
		NumDataRecord(string(numdatar)),
		Duration(string(duration)),
		NumSignal(string(numsignal)),
		Labels(label),
		TransducerTypes(transducerType),
		PhysicalDimensions(phydim),
		PhysicalMins(phymin),
		PhysicalMaxs(phymax),
		DigitalMins(digmin),
		DigitalMaxs(digmax),
		Prefilters(prefilter),
		NumSamples(numsample),
		NSReserved(nsreserved))

	if err != nil {
		return nil, err
	}

	numrecords, err := asciiToInt(numdatar)
	if err != nil {
		fmt.Errorf("serialize ascii to int failure %v, for %v\n", err, numdatar)
	}
	d := make([]*Data, numrecords)
	for i := 0; i < numrecords; i++ {
		d[i] = &Data{signals: make([][]int, ns)}
		for j := 0; j < ns; j++ {
			sampinsig, err := strToInt(numsample[j])
			if err != nil {
				fmt.Errorf("serialize ascii to int failure %v, for %v\n", err, numdatar)
			}
			bytesize, err := strToInt(nsreserved[j])
			if err != nil {
				bytesize = 2
			}
			bytesinsig := sampinsig * bytesize
			d[i].signals[j], err = ToInt(sampinsig, buf[:bytesinsig])
			if err != nil {
				fmt.Errorf("serialize bytes to int failure %v\n", err)
			}
			if len(buf) != 0 {
				buf = buf[bytesinsig:]
			}
		}
	}
	edf = NewEDF(h, d)

	return edf, nil
}

// Marshal edf into byte slice
func Marshal(edf *EDF) (buf []byte, err error) {
	buf, err = edf.header.AppendContents(buf)
	if err != nil {
		return nil, err
	}
	buf = edf.ConcatDataRecords(buf)
	return buf, nil
}

func NewEDF(h *Header, d []*Data) *EDF {
	return &EDF{
		header:      h,
		dataRecords: d,
	}
}

type EDF struct {
	header      *Header
	dataRecords []*Data
}

func (e *EDF) ConcatDataRecords(buf []byte) []byte {
	for _, record := range e.dataRecords {
		buf = append(buf, record.rawData...)
	}
	return buf
}

// NewData ...
func NewData(rawData []byte) *Data {
	return &Data{
		rawData: rawData,
	}
}

// Data holds edf data record
type Data struct {
	rawData []byte
	signals [][]int
}

// NewHeader instantiates a edf header
// Optional funcational parameters to set values in header
// If optional parameters are left blank, default is ASCII NUL
func NewHeader(options ...func(*Header) error) (*Header, error) {
	h := &Header{}
	h.allocateFixed()
	for _, option := range options {
		if err := option(h); err != nil {
			return nil, err
		}
	}
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return h, err
	}
	nb := h.calcNumBytes(ns)
	h.setNumBytes(strconv.Itoa(nb))
	h.allocateVariable(ns)
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
	numsignal      [4]byte
	label          [][16]byte
	transducerType [][80]byte
	phydim         [][8]byte
	phymin         [][8]byte
	phymax         [][8]byte
	digmin         [][8]byte
	digmax         [][8]byte
	prefilter      [][80]byte
	numsample      [][8]byte
	nsreserved     [][32]byte
}

func (h *Header) setVersion(number string) error {
	var idl int
	for idx, val := range number {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setVersion\n", errNotPrintable, val)
		}
		h.version[idx] = number[idx]
		idl = idx
	}
	fillWithSpaces(h.version[idl+1:])
	return nil
}

func (h *Header) setLPID(localPatientID string) error {
	var idl int
	for idx, val := range localPatientID {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setLPID\n", errNotPrintable, val)
		}
		h.LPID[idx] = localPatientID[idx]
		idl = idx
	}
	fillWithSpaces(h.LPID[idl+1:])
	return nil
}

func (h *Header) setLRID(localRecordID string) error {
	var idl int
	for idx, val := range localRecordID {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setLRID\n", errNotPrintable, val)
		}
		h.LRID[idx] = localRecordID[idx]
		idl = idx
	}
	fillWithSpaces(h.LRID[idl+1:])
	return nil
}

func (h *Header) setStartdate(startdate string) error {
	var idl int
	for idx, val := range startdate {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setStartdate\n", errNotPrintable, val)
		}
		h.startdate[idx] = startdate[idx]
		idl = idx
	}
	fillWithSpaces(h.startdate[idl+1:])
	return nil
}

func (h *Header) setStarttime(starttime string) error {
	var idl int
	for idx, val := range starttime {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setStarttime\n", errNotPrintable, val)
		}
		h.starttime[idx] = starttime[idx]
		idl = idx
	}
	fillWithSpaces(h.starttime[idl+1:])
	return nil
}

func (h *Header) setNumBytes(numbytes string) error {
	var idl int
	for idx, val := range numbytes {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setNumBytes\n", errNotPrintable, val)
		}
		h.numbytes[idx] = numbytes[idx]
		idl = idx
	}
	fillWithSpaces(h.numbytes[idl+1:])
	return nil
}

func (h *Header) setNumDataRecord(numdatar string) error {
	var idl int
	for idx, val := range numdatar {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setNumDataRecord\n", errNotPrintable, val)
		}
		h.numdatar[idx] = numdatar[idx]
		idl = idx
	}
	fillWithSpaces(h.numdatar[idl+1:])
	return nil
}

func (h *Header) setDuration(dur string) error {
	var idl int
	for idx, val := range dur {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setDuration\n", errNotPrintable, val)
		}
		h.duration[idx] = dur[idx]
	}
	fillWithSpaces(h.duration[idl+1:])
	return nil
}

func (h *Header) setNumSig(ns string) error {
	var idl int
	for idx, val := range ns {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setNumSig\n", errNotPrintable, val)
		}
		h.numsignal[idx] = ns[idx]
		idl = idx
	}
	fillWithSpaces(h.numsignal[idl+1:])
	return nil
}

func (h *Header) setReserved(res string) error {
	var idl int
	for idx, val := range res {
		if val < 32 || val > 126 {
			return fmt.Errorf("%s for %v in setReserved\n", errNotPrintable, val)
		}
		h.reserved[idx] = res[idx]
		idl = idx
	}
	if idl == 0 {
		fillWithSpaces(h.reserved[:])
	} else {
		fillWithSpaces(h.reserved[idl+1:])
	}

	return nil
}

func (h *Header) setLabels(labels []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.label = make([][16]byte, ns)
	for idz, label := range labels {
		for idx, val := range label {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setLabels\n", errNotPrintable, val)
			}
			h.label[idz][idx] = label[idx]
			idl = idx
		}
		fillWithSpaces(h.label[idz][idl+1:])
	}
	return nil
}

func (h *Header) setTransducerTypes(tts []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.transducerType = make([][80]byte, ns)
	for idz, tt := range tts {
		for idx, val := range tt {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setTransTypes\n", errNotPrintable, val)
			}
			h.transducerType[idz][idx] = tt[idx]
			idl = idx
		}
		fillWithSpaces(h.transducerType[idz][idl+1:])
	}
	return nil
}

func (h *Header) setPhysicalDimensions(phydims []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.phydim = make([][8]byte, ns)
	for idz, phydim := range phydims {
		for idx, val := range phydim {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setPhyDim\n", errNotPrintable, val)
			}
			h.phydim[idz][idx] = phydim[idx]
			idl = idx
		}
		fillWithSpaces(h.phydim[idz][idl+1:])
	}
	return nil
}

func (h *Header) setPhysicalMins(phymins []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.phymin = make([][8]byte, ns)
	for idz, phymin := range phymins {
		for idx, val := range phymin {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setPhyMin\n", errNotPrintable, val)
			}
			h.phymin[idz][idx] = phymin[idx]
			idl = idx
		}
		fillWithSpaces(h.phymin[idz][idl+1:])
	}
	return nil
}

func (h *Header) setPhysicalMaxs(phymaxs []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.phymax = make([][8]byte, ns)
	for idz, phymax := range phymaxs {
		for idx, val := range phymax {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setPhyMax\n", errNotPrintable, val)
			}
			h.phymax[idz][idx] = phymax[idx]
			idl = idx
		}
		fillWithSpaces(h.phymax[idz][idl+1:])
	}
	return nil
}

func (h *Header) setDigitalMins(digmins []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.digmin = make([][8]byte, ns)
	for idz, digmin := range digmins {
		for idx, val := range digmin {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setDigMin\n", errNotPrintable, val)
			}
			h.digmin[idz][idx] = digmin[idx]
			idl = idx
		}
		fillWithSpaces(h.digmin[idz][idl+1:])
	}
	return nil
}

func (h *Header) setDigitalMaxs(digmaxs []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.digmax = make([][8]byte, ns)
	for idz, digmax := range digmaxs {
		for idx, val := range digmax {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setDigMax\n", errNotPrintable, val)
			}
			h.digmax[idz][idx] = digmax[idx]
			idl = idx
		}
		fillWithSpaces(h.digmax[idz][idl+1:])
	}
	return nil
}

func (h *Header) setPrefilters(prefilters []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.prefilter = make([][80]byte, ns)
	for idz, prefilter := range prefilters {
		for idx, val := range prefilter {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setPrefilters\n", errNotPrintable, val)
			}
			h.prefilter[idz][idx] = prefilter[idx]
			idl = idx
		}
		fillWithSpaces(h.prefilter[idz][idl+1:])
	}
	return nil
}

func (h *Header) setNumSamples(numsamples []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.numsample = make([][8]byte, ns)
	for idz, numsample := range numsamples {
		for idx, val := range numsample {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setNumSamples\n", errNotPrintable, val)
			}
			h.numsample[idz][idx] = numsample[idx]
			idl = idx
		}
		fillWithSpaces(h.numsample[idz][idl+1:])
	}
	return nil
}

func (h *Header) setNSReserved(nsreserved []string) error {
	var idl int
	ns, err := asciiToInt(h.numsignal[:])
	if err != nil {
		return err
	}
	h.nsreserved = make([][32]byte, ns)
	for idz, nsres := range nsreserved {
		for idx, val := range nsres {
			if val < 32 || val > 126 {
				return fmt.Errorf("%s for %v in setNSReserved\n", errNotPrintable, val)
			}
			h.nsreserved[idz][idx] = nsres[idx]
			idl = idx
		}
		if idl == 0 {
			fillWithSpaces(h.nsreserved[idz][:])
		} else {
			fillWithSpaces(h.nsreserved[idz][idl+1:])
		}
	}
	return nil
}
func (h *Header) allocateFixed() {
	fillWithSpaces(h.version[:])
	fillWithSpaces(h.LPID[:])
	fillWithSpaces(h.LRID[:])
	fillWithSpaces(h.startdate[:])
	fillWithSpaces(h.starttime[:])
	fillWithSpaces(h.numbytes[:])
	fillWithSpaces(h.reserved[:])
	fillWithSpaces(h.numdatar[:])
	fillWithSpaces(h.duration[:])
	fillWithSpaces(h.numsignal[:])
}

func (h *Header) allocateVariable(ns int) {
	if len(h.label) == 0 {
		h.label = make([][len(h.label[0])]byte, ns)
		for _, label := range h.label {
			fillWithSpaces(label[:])
		}
	}
	if len(h.transducerType) == 0 {
		h.transducerType = make([][len(h.transducerType[0])]byte, ns)
		for _, transducerType := range h.transducerType {
			fillWithSpaces(transducerType[:])
		}
	}
	if len(h.phydim) == 0 {
		h.phydim = make([][len(h.phydim[0])]byte, ns)
		for _, phydim := range h.phydim {
			fillWithSpaces(phydim[:])
		}
	}
	if len(h.phymin) == 0 {
		h.phymin = make([][len(h.phymin[0])]byte, ns)
		for _, phymin := range h.phymin {
			fillWithSpaces(phymin[:])
		}
	}
	if len(h.phymax) == 0 {
		h.phymax = make([][len(h.phymax[0])]byte, ns)
		for _, phymax := range h.phymax {
			fillWithSpaces(phymax[:])
		}
	}
	if len(h.digmin) == 0 {
		h.digmin = make([][len(h.digmin[0])]byte, ns)
		for _, digmin := range h.digmin {
			fillWithSpaces(digmin[:])
		}
	}
	if len(h.digmax) == 0 {
		h.digmax = make([][len(h.digmax[0])]byte, ns)
		for _, digmax := range h.digmax {
			fillWithSpaces(digmax[:])
		}
	}
	if len(h.prefilter) == 0 {
		h.prefilter = make([][len(h.prefilter[0])]byte, ns)
		for _, prefilter := range h.prefilter {
			fillWithSpaces(prefilter[:])
		}
	}
	if len(h.numsample) == 0 {
		h.numsample = make([][len(h.numsample[0])]byte, ns)
		for _, numsample := range h.numsample {
			fillWithSpaces(numsample[:])
		}
	}
	if len(h.nsreserved) == 0 {
		h.nsreserved = make([][len(h.nsreserved[0])]byte, ns)
		for _, nsreserved := range h.nsreserved {
			fillWithSpaces(nsreserved[:])
		}
	}
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
	contents = append(contents, h.numsignal[:]...)
	ns, err := asciiToInt(h.numsignal[:])
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
	nb += len(h.numsignal)
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

// NumBytes setter
func NumBytes(numbytes string) func(*Header) error {
	return func(h *Header) error {
		return h.setNumBytes(numbytes)
	}
}

// NumDataR setter
func NumDataRecord(nr string) func(*Header) error {
	return func(h *Header) error {
		return h.setNumDataRecord(nr)
	}
}

// Duration setter
func Duration(dur string) func(*Header) error {
	return func(h *Header) error {
		return h.setDuration(dur)
	}
}

// NumSig setter
func NumSignal(ns string) func(*Header) error {
	return func(h *Header) error {
		return h.setNumSig(ns)
	}
}

// Reserved setter
func Reserved(res string) func(*Header) error {
	return func(h *Header) error {
		return h.setReserved(res)
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

// NSReserved setter
func NSReserved(nsreserved []string) func(*Header) error {
	return func(h *Header) error {
		return h.setNSReserved(nsreserved)
	}
}
