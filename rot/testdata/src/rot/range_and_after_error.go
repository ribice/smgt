package rot

import (
	"log"
	"os"
	"sync"
)

func testRangeVariable() {
	keys := []string{"a", "b", "c"}
	merge := map[string]int{"a": 1, "b": 2, "c": 3}
	for idx, key := range keys {
		if idx > 0 {
			log.Print(", ")
		}
		log.Printf("%s=%d", key, merge[key])
	}
}

func testChannelAfterErrorCheck() {
	smsChannelSize := 10
	csvOutputPath := "test.csv"
	csvOutputFileMode := os.FileMode(0644)
	headers := []string{"header1", "header2"}
	var db interface{}
	var logger interface{}

	smsCh := make(chan int, smsChannelSize)
	file, err := os.OpenFile(csvOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, csvOutputFileMode)
	if err != nil {
		logger = err
		return
	}
	writer := &csvWriter{}
	if err := writer.Write(headers); err != nil {
		logger = err
		return
	}
	wg := sync.WaitGroup{}
	go processSMSRecord(smsCh, &wg, writer, db, logger)
	_ = file
}

type csvWriter struct{}

func (w *csvWriter) Write(headers []string) error {
	return nil
}

func processSMSRecord(ch chan int, wg *sync.WaitGroup, writer *csvWriter, db interface{}, logger interface{}) {
}

