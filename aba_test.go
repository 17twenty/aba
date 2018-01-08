package aba

import (
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
