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
			NameOfRemitter:  "SpaceshipAU",
		},
		{
			AccountNumber:      "12112",
			BSBNumber:          "999-888",
			TransactionCode:    Credit,
			Title:              "MR NICK GLYNN",
			TraceBSB:           "999-999",
			TraceAccount:       "999999999",
			NameOfRemitter:     "SpaceshipAU",
			LodgementReference: "SuperstarHeroBUTTHISBITMAKESITTOOLONGOMGWTFBBQ",
		},
		{},
	}

	w := NewWriter(os.Stdout)

	w.Description = "WeeklyDebit"
	w.NameOfUserID = "Macquarie Bank LTD"
	w.APCAUserID = 181
	w.NameOfUsersBank = "MBL"

	if err := w.Write(records); err != nil {
		t.Fatal("error writing record to aba:", err)
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Println(err)
	}
}

func TestMinimumLength(t *testing.T) {
	records := []Record{
		{},
	}

	w := NewWriter(os.Stdout)

	if err := w.Write(records); err != ErrInsufficientRecords {
		t.Fatal("Expected '", ErrInsufficientRecords, "' but got", err)
	}

}

// 0        01MBL        Macquarie Bank LTD       999181Creditors    110811XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
