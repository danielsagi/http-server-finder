package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

type MatchStatus bool

func checkServer(url, method, headerKey string, valueRegex *regexp.Regexp, httpClient *http.Client, output chan<- string) (ms MatchStatus) {
	ms = MatchStatus(false)
	req, err := http.NewRequest(method, url, nil)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Checing", url)
	if val, ok := resp.Header[headerKey]; ok {
		joinedHeaders := []byte(strings.Join(val, " "))
		if MatchStatus(valueRegex.Match(joinedHeaders)) {
			output <- fmt.Sprintf("%s:%s", url, joinedHeaders)
		}
	}

	output <- ""
	return
}

func main() {
	var opts struct {
		Verbose     []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
		Timeout     int    `short:"t" long:"timeout" description:"HTTP Timeout"`
		Method      string `short:"X" long:"method" description:"HTTP Method to use" required:"true"`
		File        string `short:"w" long:"targets-file" description:"Path to a newline seperated targets" required:"true"`
		HeaderName  string `short:"k" long:"header-key" description:"Response Header Key To Match"  required:"true"`
		HeaderRegex string `short:"r" long:"regex-value" description:"Response header value regex to match for" required:"true"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	// compiling regex
	regExpr, _ := regexp.Compile(opts.HeaderRegex)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c := http.Client{Timeout: time.Duration(opts.Timeout) * time.Second}

	targetCounter := 0
	outputStatusesChannel := make(chan string)

	// reading targets file
	file, err := os.Open(opts.File)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// For each target in file create a goroutine
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currTarget := scanner.Text()
		go checkServer(currTarget, opts.Method, opts.HeaderName, regExpr, &c, outputStatusesChannel)
		targetCounter++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// collect success statuses and print
	for targetCounter > 0 {
		fmt.Printf("%d\n", targetCounter)
		status := <-outputStatusesChannel
		if status != "" {
			fmt.Println("Matched:", status)
		}
		targetCounter--
	}
}
