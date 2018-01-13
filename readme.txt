PACKAGE DOCUMENTATION

package aba
    import "."


CONSTANTS

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

VARIABLES

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
)

TYPES

type Reader struct {
    Header  header
    Trailer trailer
    // contains filtered or unexported fields
}
    A Reader reads records from an ABA file.

    As returned by NewReader, a Reader expects input conforming to spec. The
    Header and Trailer fields expose details about the underlying item

func NewReader(r io.Reader) *Reader
    NewReader returns a new Reader that reads from r.

func (r *Reader) Read() (record Record, err error)
    Read reads one record (a slice of fields) from r First read may not
    return a record, instead Header will likely be populated. You may also
    get Trailer as the last

func (r *Reader) ReadAll() (records []Record, err error)
    ReadAll reads all the remaining records from r.

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
    Record is the type we care about for writing to file

func (r *Record) IsValid() bool
    IsValid performs some basic checks on records

func (r *Record) Read(l string) error

func (r *Record) Write(w io.Writer)

type Writer struct {
    // OmitBatchTotals can be used for banks that don't summarise
    // the credit/debit transactions
    OmitBatchTotals bool
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
    Write writes the provided []Records into the buffer. It returns an error
    if something is wrong with the records.


