package aba

import (
	"bufio"
	"io"
	"log"
)

// A Reader reads records from an ABA file.
//
// As returned by NewReader, a Reader expects input conforming to spec.
// The Header and Trailer fields expose details about the underlying item
type Reader struct {
	Header  *Header
	Trailer *Trailer
	r       *bufio.Reader
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:       bufio.NewReader(r),
		Header:  &Header{},
		Trailer: &Trailer{},
	}
}

// Read reads one record (a slice of fields) from r
// First read may not return a record, instead Header
// will likely be populated. You may also get Trailer
// as the last
func (r *Reader) Read() (record Record, err error) {
	return r.readRecordOrHeaderOrTrailer()
}

// ReadAll reads all the remaining records from r.
func (r *Reader) ReadAll() (records []Record, err error) {

	for {
		var rec Record
		rec, err = r.readRecordOrHeaderOrTrailer()
		if err == io.EOF {
			err = nil // ReadAll is happy - not erroneous
			return
		}
		if err != nil {
			log.Println("readRecordOrHeaderOrTrailer", err)
			break
		}
		// No point appending garbage
		if rec.IsValid() {
			records = append(records, rec)
		}
	}
	return
}

func (r *Reader) readRecordOrHeaderOrTrailer() (Record, error) {
	var record Record
	b, err := r.r.ReadByte()
	if err != nil || r.r.UnreadByte() != nil {
		return record, err
	}

	// We'll always want a line
	line, err := r.r.ReadString('\n')
	if err != nil && err != io.EOF {
		// Could be a trailer - there's no newline there. Look for EOF?
		log.Println("Didn't get a line")
		return record, err
	}

	switch b {
	case '0':
		err = r.Header.Read(line)

	case '1':
		err = record.Read(line)
	case '7':
		err = r.Trailer.Read(line)
	default:
		err = ErrUnexpectedRecordType
	}

	return record, err
}
