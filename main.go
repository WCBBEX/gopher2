package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func checkWebsite(url string, client *http.Client) bool {
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()
	return true
}

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func main() {
	now := time.Now()

	dataFile, err := os.Open("data.csv")
	if err != nil {
		panic(err)
	}
	defer dataFile.Close()

	reader := csv.NewReader(dataFile)
	head, err := reader.Read()
	if err != nil {
		panic(err)
	}

	goodFile, err := os.Create("good.csv")
	if err != nil {
		panic(err)
	}
	defer goodFile.Close()
	goodWriter := csv.NewWriter(goodFile)
	defer goodWriter.Flush()

	badFile, err := os.Create("bad.csv")
	if err != nil {
		panic(err)
	}
	defer badFile.Close()
	badWriter := csv.NewWriter(badFile)
	defer badWriter.Flush()

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup
	good := make(chan []string, 100)
	bad := make(chan []string, 100)

	wg2.Add(1)
	go func() {
		defer wg2.Done()
		err := goodWriter.Write(head)
		if err != nil {
			panic(err)
		}
		for record := range good {
			err := goodWriter.Write(record)
			if err != nil {
				panic(err)
			}
		}
	}()

	wg2.Add(1)
	go func() {
		defer wg2.Done()
		err := badWriter.Write(head)
		if err != nil {
			panic(err)
		}
		for record := range bad {
			err := badWriter.Write(record)
			if err != nil {
				panic(err)
			}
		}
	}()

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				wg1.Wait()
				close(good)
				close(bad)
				break
			}
			panic(err)
		}

		wg1.Add(1)
		go func(url string) {
			defer wg1.Done()

			if checkWebsite(url, client) {
				good <- record
			} else {
				bad <- record
			}
		}(record[4])
	}
	wg2.Wait()

	fmt.Println(time.Since(now))
}
