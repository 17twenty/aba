PACKAGE DOCUMENTATION

package aba
    import "git.s8.network/nick/aba.git"


CONSTANTS

const (
    Credit = "50"
    Debit  = "13"
)

VARIABLES

var (
    // ErrInsufficientRecords is returned when the minimum 2 records is not provided for writing
    ErrInsufficientRecords   = errors.New("aba: Not enough records (minimum 2 required)")
    ErrMustSpecifyUsersBank  = errors.New("aba: Didn't specify the users bank")
    ErrMustSpecifyUsersID    = errors.New("aba: Didn't specify the users ID")
    ErrMustSpecifyAPCAUserID = errors.New("aba: Didn't specify the APCA ID")
)

TYPES

type Record struct {
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
    // contains filtered or unexported fields
}
    Record is the type we care about for writing to file

func (r *Record) IsValid() bool
    IsValid performs some basic checks on records

func (r *Record) Write(w io.Writer)

type Writer struct {
    // contains filtered or unexported fields
}
    Writer implements buffering for an io.Writer object. If an error occurs
    writing to a Writer, no more data will be accepted and all subsequent
    writes, and Flush, will return the error. After all data has been
    written, the client should call the Flush method to guarantee all data
    has been forwarded to the underlying io.Writer.

func NewWriter(w io.Writer) *Writer
    NewWriter returns a new Writer whose buffer has the default size.

func (w *Writer) Error() error
    Error reports any error that has occurred during a previous Write or
    Flush.

func (w *Writer) Flush()
    Flush can be called to ensure all data has been written

func (w *Writer) Write(records []Record) (err error)
    Write writes the contents of p into the buffer. It returns the number of
    bytes written. If nn < len(p), it also returns an error explaining why
    the write is short.
