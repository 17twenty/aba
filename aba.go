package aba

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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
	ErrBadHeader             = errors.New("aba: Bad Header prevented reading")
	ErrBadRecord             = errors.New("aba: Bad Record prevented reading")
	ErrBadTrailer            = errors.New("aba: Bad Trailer prevented reading")
	ErrUnexpectedRecordType  = errors.New("aba: Unexpected record type, can decode 0,1 and 7 only")

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

// asciiSafe replaces any non-printable ASCII chars with a '.'
func asciiSafe(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	safeString, _, _ := transform.String(t, s)
	var resultString string
	for _, r := range safeString {
		// ' ' and '~' are the endpoints of the printable ASCII range.
		if !(r < ' ' || r > '~') {
			resultString += fmt.Sprintf("%c", r)
		}
	}
	return resultString
}

type Header struct {
	RecordType         int       // pos 1 - always zero
	FileSequenceNumber int       // pos 19-20 - right justified, can increment or set to 01
	NameOfUsersBank    string    // pos 21-23 - always MBL
	NameOfUserID       string    // pos 31-56 - left justified/blank filled
	APCAUserID         int       // pos 57-62 - right justified/zero filled
	Description        string    // pos 63-74 - left justified/blank filled e.g. "     rent collection"
	ProcessingDate     time.Time // pos 75-80 DDMMYYYY and zero filled
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

func (h *Header) Write(w io.Writer) {
	tempStr := fmt.Sprintf(
		"%d%s%2.2s%3.3s%s%26.26s%06.6d%12.12s%v",
		h.RecordType,
		spaces(17),
		fmt.Sprintf("%02d", h.FileSequenceNumber),
		h.NameOfUsersBank,
		spaces(7), // 7 BlankSpaces
		padRight(asciiSafe(h.NameOfUserID), " ", 26),
		h.APCAUserID,
		padRight(asciiSafe(h.Description), " ", 12),
		h.ProcessingDate.Format("020106"),
	)
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, " ", 120))
}

func (h *Header) Read(l string) error {
	if len(l) != 121 && len(l) != 122 { // 120 + '\n' || 120 + '\r\n'
		log.Println("ABA: Header expected 120, got", len(l))
		return ErrBadHeader
	}
	// Just read it all back in and unpack
	h.RecordType, _ = strconv.Atoi(strings.TrimSpace(l[0:1]))
	h.FileSequenceNumber, _ = strconv.Atoi(strings.TrimSpace(l[18:20]))
	h.NameOfUsersBank = strings.TrimSpace(l[20:23])
	h.NameOfUserID = strings.TrimSpace(l[30:56])
	h.APCAUserID, _ = strconv.Atoi(strings.TrimSpace(l[56:62]))
	h.Description = strings.TrimSpace(l[62:74])
	h.ProcessingDate, _ = time.Parse("020106", strings.TrimSpace(l[74:80]))

	return nil
}

// Record is the type we care about for writing to file
type Record struct {
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
	NameOfRemitter         string // pos 97-112 - name of originator which may differ from username
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
		fallthrough
	case Pay:
		// All good - next checks
	default:
		return false
	}

	// Title validation - can't be all blank
	if len(strings.TrimSpace(r.Title)) == 0 {
		log.Println("Title validation")
		return false
	}

	// BSB validation
	if !bsbNumberRegEx.MatchString(r.TraceBSB) {
		log.Println("Bad TraceBSB")
		return false
	}
	if !bsbNumberRegEx.MatchString(r.BSBNumber) {
		log.Println("Bad BSB")
		return false
	}
	return true
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
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, "#", 120))
}

func (r *Record) Read(l string) error {
	if len(l) != 121 && len(l) != 122 { // 120 + '\n' || 120 + '\r\n'
		return ErrBadRecord
	}
	// Just read it all back in and unpack
	r.BSBNumber = strings.TrimSpace(l[1:8])
	r.AccountNumber = strings.TrimSpace(l[8:17])
	r.Indicator = strings.TrimSpace(l[17:18])
	r.TransactionCode = strings.TrimSpace(l[18:20])
	t, _ := strconv.Atoi(strings.TrimSpace(l[20:30]))
	r.Amount = uint64(t)
	r.Title = strings.TrimSpace(l[30:62])
	r.LodgementReference = strings.TrimSpace(l[62:80])
	r.TraceBSB = strings.TrimSpace(l[80:87])
	r.TraceAccount = strings.TrimSpace(l[87:96])
	r.NameOfRemitter = strings.TrimSpace(l[96:112])
	t, _ = strconv.Atoi(strings.TrimSpace(l[112:120]))
	r.AmountOfWithholdingTax = uint64(t)

	if !r.IsValid() {
		return ErrBadRecord
	}
	return nil
}

type Trailer struct {
	RecordType         int    // pos 1 - always seven
	DefaultBSB         string // pos 2-8 - always 999-999
	UserNetTotalAmount uint64 // pos 21-30 - Right justfied in cents without punctuation i.e 0000000000
	// in a balanced file, the credit and debit total should always be the same
	UserCreditTotalAmount uint64 // pos 31-40 - Right justified in cents e,g, $100.00 == 10000
	UserDebitTotalAmount  uint64 // pos 41-50 - Right justified in cents e,g, $100.00 == 10000
	UserTotalRecords      int    // pos 75-80 - Right Justified of size 6
	// Space filled from 81-120. Spaces between every gap for a total 120 characters
}

func (t *Trailer) Write(w io.Writer) {
	tempStr := fmt.Sprintf(
		"%d%.7s%s%010.10d%010.10d%010.10d%s%06d%s",
		t.RecordType,
		t.DefaultBSB,
		spaces(12),
		t.UserNetTotalAmount,
		t.UserCreditTotalAmount,
		t.UserDebitTotalAmount,
		spaces(24),
		t.UserTotalRecords,
		spaces(40),
	)
	// Add final padding
	fmt.Fprintf(w, "%s", padRight(tempStr, " ", 120))
}

func (t *Trailer) Read(l string) error {
	if len(l) != 120 { // 120 and no newline
		log.Println("ABA: Trailer expected 120, got", len(l))
		return ErrBadTrailer
	}
	// Just read it all back in and unpack
	t.RecordType, _ = strconv.Atoi(strings.TrimSpace(l[0:1]))

	t.DefaultBSB = strings.TrimSpace(l[1:8])

	tmp, _ := strconv.Atoi(strings.TrimSpace(l[20:30]))
	t.UserNetTotalAmount = uint64(tmp)

	tmp, _ = strconv.Atoi(strings.TrimSpace(l[30:40]))
	t.UserCreditTotalAmount = uint64(tmp)

	tmp, _ = strconv.Atoi(strings.TrimSpace(l[40:50]))
	t.UserDebitTotalAmount = uint64(tmp)
	t.UserTotalRecords, _ = strconv.Atoi(strings.TrimSpace(l[74:80]))

	return nil
}
