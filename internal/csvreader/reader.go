package csvreader

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

// Read and parse csv file to slice of records.
func Read(filePath string) ([][]string, error) {
	// Open the file
	csvfile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	records := [][]string{}
	// Parse the file
	r := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) < 2 {
			log.Println("records length less then 2")

			continue
		}

		fmt.Printf("Product: %s Price %s\n", record[0], record[1])
		records = append(records, record)
	}

	return records, nil
}
