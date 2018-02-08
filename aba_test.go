package aba

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestDemo(t *testing.T) {
	records := []Record{
		{
			AccountNumber:   "3424",
			BSBNumber:       "888-123",
			TransactionCode: Credit,
			Title:           "DEMO DEMO",
			TraceBSB:        "111-111",
			TraceAccount:    "999999999",
			Amount:          1000,
			NameOfRemitter:  "SpaceshipAU",
		},
		{
			AccountNumber:      "12112",
			BSBNumber:          "999-888",
			TransactionCode:    Credit,
			Title:              "MR NICK GLYNN",
			Amount:             1000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "SpaceshipAU",
			LodgementReference: "SuperstarHeroBUTTHISBITMAKESITTOOLONGOMGWTFBBQ",
		},
		{
			AccountNumber:      "260070750",
			BSBNumber:          "182-222",
			TransactionCode:    Debit,
			Title:              "Macquarite Account",
			Amount:             2000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "ddu",
			LodgementReference: "ABLE",
		},
	}

	w := NewWriter(os.Stdout)

	w.Description = "WeeklyDebit"
	w.NameOfUserID = "Macquarie Bank LTD"
	w.APCAUserID = 181
	w.NameOfUsersBank = "MBL"

	if err := w.Write(records); err != nil {
		t.Fatal("error writing record", err)
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Println(err)
	}
}

func TestMinimumLength(t *testing.T) {
	records := []Record{}

	w := NewWriter(os.Stdout)

	if err := w.Write(records); err != ErrInsufficientRecords {
		t.Fatal("Expected '", ErrInsufficientRecords, "' but got", err)
	}

}

func TestLocalReader(t *testing.T) {
	f, err := os.Open("./Test_ABA_20180108.aba")

	if err != nil {
		t.Fatal("Couldn't find local test file")
	}

	nr := NewReader(f)
	r, err := nr.ReadAll()

	if err != nil {
		t.Fatal("Expected '", nil, "' but got", err)
	}

	if len(r) != 4 && nr.Trailer.userTotalRecords != 4 {
		t.Fatalf("Failure - expected 3 records but got %v and %v\n", len(r), nr.Trailer.userTotalRecords)
	}

	for idx, val := range []string{
		"Paul Smith",
		"John Dickson",
		"Peter Jackson",
		"Sacha Belle",
	} {
		if val != r[idx].Title {
			t.Fatal("Expected '", val, "' but got", r[idx].Title)
		}
	}
}

func TestLocalWriter(t *testing.T) {
	f, err := os.Create("./DELETE_ME")

	if err != nil {
		t.Fatal("Couldn't open local test file")
	}

	records := []Record{
		{
			AccountNumber:   "3424",
			BSBNumber:       "888-123",
			TransactionCode: Credit,
			Title:           "DEMO DEMO",
			TraceBSB:        "111-111",
			TraceAccount:    "999999999",
			Amount:          1000,
			NameOfRemitter:  "SpaceshipAU",
		},
		{
			AccountNumber:      "12112",
			BSBNumber:          "999-888",
			TransactionCode:    Credit,
			Title:              "MR NICK GLYNN",
			Amount:             1000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "SpaceshipAU",
			LodgementReference: "SuperstarHeroBUTTHISBITMAKESITTOOLONGOMGWTFBBQ",
		},
		{
			AccountNumber:      "260070750",
			BSBNumber:          "182-222",
			TransactionCode:    Debit,
			Title:              "Macquarite Account",
			Amount:             2000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "ddu",
			LodgementReference: "ABLE",
		},
	}

	w := NewWriter(f)

	// To handle Windows encoding... urgh
	w.CRLFLineEndings = true

	w.Description = "WeeklyDebit"
	w.NameOfUserID = "Macquarie Bank LTD"
	w.APCAUserID = 181
	w.NameOfUsersBank = "MBL"

	if err := w.Write(records); err != nil {
		t.Fatal("error writing record", err)
	}
	w.Flush()
}

func TestWriteReader(t *testing.T) {

	records := []Record{
		{
			AccountNumber:   "3424",
			BSBNumber:       "888-123",
			TransactionCode: Credit,
			Title:           "DEMO DEMO",
			TraceBSB:        "111-111",
			TraceAccount:    "999999999",
			Amount:          1000,
			NameOfRemitter:  "SpaceshipAU",
		},
		{
			AccountNumber:      "12112",
			BSBNumber:          "999-888",
			TransactionCode:    Credit,
			Title:              "MR NICK GLYNN",
			Amount:             1000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "SpaceshipAU",
			LodgementReference: "SuperstarHeroBUTTHISBITMAKESITTOOLONGOMGWTFBBQ",
		},
		{
			AccountNumber:      "260070750",
			BSBNumber:          "182-222",
			TransactionCode:    Debit,
			Title:              "Macquarite Account",
			Amount:             2000,
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "ddu",
			LodgementReference: "ABLE",
		},
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)

	w.Description = "WeeklyDebit"
	w.NameOfUserID = "Macquarie Bank LTD"
	w.APCAUserID = 181
	w.NameOfUsersBank = "MBL"

	if err := w.Write(records); err != nil {
		t.Fatal("error writing record", err)
	}

	w.Flush()

	log.Println("Size in buffer:", buf.Len())

	f := NewReader(&buf)
	r, err := f.ReadAll()

	if err != nil {
		t.Fatal("Expected '", nil, "' but got", err)
	}

	if len(r) != 3 && f.Trailer.userTotalRecords != 3 {
		t.Fatalf("Failure - expected 3 records but got %v and %v\n", len(r), f.Trailer.userTotalRecords)
	}
}
