package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func cleanURL(url string) string {
	return strings.TrimSuffix(url, ",")
}

func extractURLs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	urlRegex := regexp.MustCompile(`https://[^\s]+`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if urlRegex.MatchString(line) {
			urls = append(urls, urlRegex.FindString(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func getCertInfo(url string) (*tls.ConnectionState, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DisableKeepAlives:  true,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 40 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		if resp != nil && resp.TLS != nil {
			return resp.TLS, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	return resp.TLS, nil
}

func processURL(url string, wg *sync.WaitGroup, sem chan bool, records chan []string) {
	defer wg.Done()
	defer func() { <-sem }()

	cleanedURL := cleanURL(url)
	certInfo, err := getCertInfo(cleanedURL)
	if err != nil {
		fmt.Printf("Error processing URL %s: %v\n", cleanedURL, err)
		return
	}
	if certInfo != nil && len(certInfo.PeerCertificates) > 0 {
		cert := certInfo.PeerCertificates[0]
		names := map[string]bool{}

		commonName := cert.Subject.CommonName

		for _, dnsName := range cert.DNSNames {
			names[dnsName] = true
		}

		delete(names, commonName)

		var uniqueSANs []string
		for name := range names {
			uniqueSANs = append(uniqueSANs, name)
		}

		records <- []string{cleanedURL, commonName, strings.Join(uniqueSANs, "\n")}
	}
}

func printRecord(record []string) {
	fmt.Printf("URL: %s\nCommon Name: %s\nSANs:\n%s\n\n", record[0], record[1], record[2])
}

func main() {
	inputFile := flag.String("input", "", "Input file with URLs")
	outputFile := flag.String("output", "", "Output file for saving results")
	singleURL := flag.String("url", "", "Single URL to process")
	threads := flag.Int("threads", 5, "Number of concurrent threads")

	flag.Parse()

	var urls []string
	var err error

	if *singleURL != "" {
		urls = append(urls, *singleURL)
	} else {
		if *inputFile == "" {
			fmt.Println("Please specify an input file using -input=<filename> or a single URL using -url=<URL>")
			return
		}
		urls, err = extractURLs(*inputFile)
		if err != nil {
			fmt.Printf("Error reading URLs: %v\n", err)
			return
		}
	}

	var wg sync.WaitGroup
	sem := make(chan bool, *threads)
	records := make(chan []string, 10)

	go func() {
		for _, url := range urls {
			wg.Add(1)
			sem <- true
			go processURL(url, &wg, sem, records)
		}
		wg.Wait()
		close(records)
	}()

	var w *csv.Writer

	if *outputFile == "" {
		w = csv.NewWriter(os.Stdout)
	} else {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer file.Close()
		w = csv.NewWriter(file)
	}

	for record := range records {
		if err := w.Write(record); err != nil {
			fmt.Printf("Error writing record to csv: %v\n", err)
		} else {
			printRecord(record)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Printf("Error writing csv: %v\n", err)
	}

	fmt.Printf("Processing complete. %d URLs processed.\n", len(urls))
}
