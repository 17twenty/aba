package aba

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	Debit           = "13" // Externally initiated debit item
	Credit          = "50" // Initiated externally
	AGSI            = "51" // Australian Govt Security Interest
	FamilyAllowance = "52" // Family allowance
	Pay             = "53" // Pay
	Pension         = "54" // Pension
	Allotment       = "55" // Allotment
	Dividend        = "56" // Dividend
	NoteInterest    = "57" // Debenture/Note Interest
)

var (
	ErrInsufficientRecords   = errors.New("aba: Not enough records (minimum 2 required)")
	ErrMustSpecifyUsersBank  = errors.New("aba: Didn't specify the users bank")
	ErrMustSpecifyUsersID    = errors.New("aba: Didn't specify the users ID")
	ErrMustSpecifyAPCAUserID = errors.New("aba: Didn't specify the APCA ID")
	ErrInvalidRecord         = errors.New("aba: Invalid Record can't be written")

	bsbNumberRegEx = regexp.MustCompile(`^\d{3}-\d{3}$`)
)

func padRight(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}

func spaces(howMany int) string {
	return padRight("", " ", howMany)
}

type header struct {
	recordType         int       // pos 1 - always zero
	fileSequenceNumber int       // pos 19-29 - right justified, can increment or set to 01
	NameOfUsersBank    string    // pos 21-23 - always MBL
	NameOfUserID       string    // pos 31-56 - left justified/blank filled
	APCAUserID         int       // pos 57-62 - right justified/zero filled
	Description        string    // pos 63-74 - left justified/blank filled e.g. "     rent collection"
	ProcessingDate     time.Time // pos 75-80 DDMMYYYY and zero filled
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

func (h *header) Write(w io.Writer) {
	tempStr := fmt.Sprintf(
		"%d%s%2.2s%3.3s%s%26.26s%06.6d%12.12s%v",
		h.recordType,
		spaces(17),
		fmt.Sprintf("%02d", h.fileSequenceNumber),
		h.NameOfUsersBank,
		spaces(7), // 7 BlankSpaces
		padRight(h.NameOfUserID, " ", 26),
		h.APCAUserID,
		padRight(h.Description, " ", 12),
		h.ProcessingDate.Format("020106"),
	)
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, " ", 120))
}

// Record is the type we care about for writing to file
type Record struct {
	// RecordType pos 1 - always one
	BSBNumber     string // pos 2-8 - in the format 999-999
	AccountNumber string // pos 9-17
	// Indicator should be one of the following
	// W - Dividend paid to a resident of a country where double tax agreement is in force
	// X - Dividend paid to a resident of any other country
	// Y - Interest paid to all non-residents -- tax withholding is to appear at 113-120
	// N - New or varied BSB/account number or name
	// Blank otherwise
	Indicator              string // pos 18
	TransactionCode        string // pos 19-20 - Either 13, debit or 50, credit.
	Amount                 uint64 // pos 21-30 - in cents, Right justified in cents e,g, $100.00 == 10000
	Title                  string // pos 31-62 - must not be blank,. Left justified and blank filled. Title of account
	LodgementReference     string // pos 63-80 - e.g invoice number/payroll etc
	TraceBSB               string // pos 81-87 - BSB number of user supplying the file in format 999-999
	TraceAccount           string // pos 88-96 - account number of user supplying file
	NameOfRemitter         string // pos 97-112 - name of originator which may differe from username
	AmountOfWithholdingTax uint64 // pos 113-120 - Shown in cents without punctuation
}

// IsValid performs some basic checks on records
func (r *Record) IsValid() bool {

	// Transaction validation
	switch r.TransactionCode {
	case Credit:
		fallthrough
	case Debit:
		// All good - next checks
	default:
		return false
	}

	// Title validation
	// 1. Can't be all blank
	if len(strings.TrimSpace(r.Title)) == 0 {
		return false
	}

	// BSB validation
	if !bsbNumberRegEx.MatchString(r.TraceBSB) {
		return false
	}
	return bsbNumberRegEx.MatchString(r.BSBNumber)
}

func (r *Record) Write(w io.Writer) {
	tempStr := fmt.Sprintf(
		"1%7.7s%9.9s%1.1s%2.2s%010.10d%32.32s%18.18s%7.7s%9.9s%16.16s%08.8d", // Record type always 1
		r.BSBNumber,
		r.AccountNumber,
		r.Indicator,
		r.TransactionCode,
		r.Amount,
		padRight(r.Title, " ", 32),
		padRight(r.LodgementReference, " ", 18),
		r.TraceBSB,
		r.TraceAccount,
		padRight(r.NameOfRemitter, " ", 16),
		r.AmountOfWithholdingTax,
	)
	log.Println("String length:", len(tempStr))
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, "#", 120))
}

type trailer struct {
	recordType         int    // pos 1 - always seven!
	DefaultBSB         string // pos 2-8 - always 999-999
	userNetTotalAmount uint64 // pos 21-30 - Right justfied in cents without punctuation i.e 0000000000
	// in a balanced file, the credit and debit total should always be the same
	userCreditTotalAmount uint64 // pos 31-40 - Right justified in cents e,g, $100.00 == 10000
	userDebitTotalAmount  uint64 // pos 41-50 - Right justified in cents e,g, $100.00 == 10000
	userTotalRecords      int    // pos 75-80 - Right Justified of size 6
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

func (t *trailer) Write(w io.Writer) {
	tempStr := fmt.Sprintf(
		"%d%.7s%s%010.10d%010.10d%010.10d%s%06d%s",
		t.recordType,
		t.DefaultBSB,
		spaces(12),
		t.userNetTotalAmount,
		t.userCreditTotalAmount,
		t.userDebitTotalAmount,
		spaces(24),
		t.userTotalRecords,
		spaces(40),
	)
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, " ", 120))
}

// Writer implements buffering for an io.Writer object.
// If an error occurs writing to a Writer, no more data will be
// accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.
type Writer struct {
	// OmitBatchTotals can be used for banks that don't summarise
	// the credit/debit transactions
	OmitBatchTotals bool
	header
	trailer
	wr *bufio.Writer
}

// NewWriter returns a new Writer whose buffer has the default size.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		wr: bufio.NewWriter(w),
		header: header{
			recordType:         0,
			fileSequenceNumber: 1,
			APCAUserID:         181,
			Description:        "Creditors",
			ProcessingDate:     time.Now(),
		},
		trailer: trailer{
			recordType: 7,
			DefaultBSB: "999-999",
		},
	}
}

// Write writes the provided []Records into the buffer.
// It returns an error if something is wrong with the records.
func (w *Writer) Write(records []Record) (err error) {
	if len(records) < 2 {
		return ErrInsufficientRecords
	}
	if len(strings.TrimSpace(w.NameOfUsersBank)) == 0 {
		return ErrMustSpecifyUsersBank
	}

	if len(strings.TrimSpace(w.NameOfUserID)) == 0 {
		return ErrMustSpecifyUsersID
	}
	if w.APCAUserID == 0 {
		return ErrMustSpecifyAPCAUserID
	}

	// Validation spin...
	w.trailer.userTotalRecords = len(records) // Count valid records
	for i, r := range records {
		if !r.IsValid() {
			return fmt.Errorf("%v (record %d)", ErrInvalidRecord, i)
		}
	}

	w.trailer.userCreditTotalAmount = 0
	w.trailer.userDebitTotalAmount = 0
	w.header.Write(w.wr)
	w.wr.WriteByte('\n')
	for i, r := range records {
		r.Write(w.wr)
		if !w.OmitBatchTotals {
			switch r.TransactionCode {
			case Debit:
				w.trailer.userDebitTotalAmount += r.Amount
			default:
				if strings.HasPrefix(r.TransactionCode, "5") {
					w.trailer.userCreditTotalAmount += r.Amount
				} else {
					log.Println("Unknown transaction type", r.TransactionCode, "in record", i)
				}
			}
		}
		w.wr.WriteByte('\n')
	}

	// Last part is to get net trailer amount
	// TODO: Some banks require a balancing line at the bottom
	// We're going to omit it unless told otherwise
	w.trailer.userNetTotalAmount = w.trailer.userCreditTotalAmount - w.trailer.userDebitTotalAmount
	w.trailer.Write(w.wr)
	return nil
}

// Flush can be called to ensure all data has been written
func (w *Writer) Flush() {
	w.wr.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w *Writer) Error() error {
	_, err := w.wr.Write(nil)
	return err
}
