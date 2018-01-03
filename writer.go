package aba

import (
	"bufio"
	"errors"
	"io"
	"time"
)

const (
	defaultBufSize = 4096
)

var (
	ErrNoHeaderSet  = errors.New("aba: No header record set")
	ErrNoTrailerSet = errors.New("aba: No trailer record set")
)

type header struct {
	RecordType         int       // pos 1 - always zero
	FileSequenceNumber int       // pos 19-29 can increment or set to 01
	NameOfUsersBank    string    // pos 21-23 always MBL
	NameOfUserID       string    // pos 31-56
	APCAUserID         string    // pos 57-62
	Description        string    // pos 63-74 e.g. rent collection
	ProcessingDate     time.Time // pos 75-80 DDMMYYYY and zero filled
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

// Record is the type we care about for writing to file
type Record struct {
	RecordType    int    // pos 1 - always zero
	BSBNumber     string // pos 2-8 - in the format 999-999
	AccountNumber string // pos 9-17
	// Indicator should be one of the following
	// W - Dvidend paid to a resident of a country where double tax agreement is in force
	// X - Dividen paid to a resident of any other country
	// Y - Interest paid to all non-residents -- tax withholding is to appear at 113-120
	// N - New or varied BSB/account number or name
	// Blank otherwise
	Indicator              string // pos 18
	TransactionCode        string // pos 19-20 - Either 13, debit or 50, credit.
	Amount                 string // pos 21-30 - in cents, Right justified in cents e,g, $100.00 == 10000
	Title                  string // pos 31-62 - must not be blank,. Left justified and blank filled. Title of account
	LodgementReference     string // pos 63-80 - e.g invoice number/payroll etc
	TraceBSB               string // pos 81-87 - BSB number of user supplying the file in format 999-999
	TraceAccount           string // pos 88-96 - account number of user supplying file
	NameOfRemitter         string // pos 97-112 - name of originator which may differe from username
	AmountOfWithholdingTax string // pos 113-120 - Shown in cents without punctuation
}

type trailer struct {
	RecordType         int    // pos 1 - always seven!
	DefaultBSB         string // pos 2-8 - always 999-999
	UserNetTotalAmount string // pos 21-30 - Right justfied in cents without punctuation i.e 0000000000
	// in a balanced file, the credit and debit total should always be the same
	UserCreditTotalAmount        string // pos 31-40 - Right justified in cents e,g, $100.00 == 10000
	UserDebitTotalAmount         string // pos 41-50 - Right justified in cents e,g, $100.00 == 10000
	UserTotalCountOfType1Records int    // pos 75-80 - Right Justified of size 6
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

// Writer implements buffering for an io.Writer object.
// If an error occurs writing to a Writer, no more data will be
// accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.
type Writer struct {
	wr *bufio.Writer
}

// NewWriter returns a new Writer whose buffer has the default size.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		wr: bufio.NewWriter(w),
	}
}

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (w *Writer) Write(record []string) (err error) {
	for n, field := range record {
		if n > 0 {
			// Add comma?
			if err := w.wr.WriteByte('\n'); err != nil {
				return err
			}
		}
		for _, r1 := range field {
			var err error
			switch r1 {
			case '"':
				_, err = w.wr.WriteString(`""`)
			case '\r':
			case '\n':
				err = w.wr.WriteByte('\n')
			default:
				_, err = w.wr.WriteRune(r1)
			}
			if err != nil {
				return err
			}
		}

		if err := w.wr.WriteByte('"'); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) Flush() {
	w.wr.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w *Writer) Error() error {
	_, err := w.wr.Write(nil)
	return err
}
