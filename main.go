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
	"github.com/schollz/progressbar/v3"
)

type Job struct {
	Url        string
	Method     string
	HeaderName string
	RegexMatch *regexp.Regexp
}

type JobResult struct {
	Url           string
	MatchedString string
	StatusCode    int
	Success       bool
}

func NewHttpClient(timeout int) (c http.Client) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c = http.Client{Timeout: time.Duration(timeout) * time.Second}
	return
}

func checkServer(url, method, headerKey string, valueRegex *regexp.Regexp, timeout int) (result JobResult) {
	client := NewHttpClient(timeout)
	req, err := http.NewRequest(method, url, nil)

	resp, err := client.Do(req)
	if err != nil {
		result = JobResult{Success: false}
		return
	}

	if val, ok := resp.Header[headerKey]; ok {
		joinedHeaders := []byte(strings.Join(val, " "))
		if valueRegex.Match(joinedHeaders) {
			result = JobResult{
				Url:           url,
				Success:       true,
				MatchedString: string(joinedHeaders),
				StatusCode:    resp.StatusCode,
			}
		}
	}
	return
}

func worker(id, timeout int, jobs <-chan Job, results chan<- JobResult) {
	for j := range jobs {
		results <- checkServer(j.Url, j.Method, j.HeaderName, j.RegexMatch, timeout)
	}
}

func main() {
	var opts struct {
		Verbose     []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
		Timeout     int    `short:"t" long:"timeout" description:"HTTP Timeout"`
		Method      string `short:"X" long:"method" description:"HTTP Method to use" required:"true"`
		File        string `short:"w" long:"targets-file" description:"Path to a newline seperated targets" required:"true"`
		OutFile     string `short:"o" long:"out-file" description:"Path to the output file" required:"true"`
		HeaderName  string `short:"k" long:"header-key" description:"Response Header Key To Match"  required:"true"`
		HeaderRegex string `short:"r" long:"regex-value" description:"Response header value regex to match for" required:"true"`
		WorkerNum   int    `short:"n" long:"worker-num" description:"Number of workers"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	// reading targets file
	file, err := os.Open(opts.File)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// getting urls
	var targetUrls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		targetUrls = append(targetUrls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	regExpr, _ := regexp.Compile(opts.HeaderRegex)

	// starting workers
	jobs := make(chan Job, len(targetUrls))
	results := make(chan JobResult, len(targetUrls))

	for w := 1; w <= opts.WorkerNum; w++ {
		go worker(w, opts.Timeout, jobs, results)
	}

	for _, url := range targetUrls {
		jobs <- Job{
			Url:        url,
			Method:     opts.Method,
			HeaderName: opts.HeaderName,
			RegexMatch: regExpr,
		}
	}

	bar := progressbar.Default(int64(len(targetUrls)))

	// Output fetching
	fo, err := os.Create(opts.OutFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	for i := 0; i <= len(targetUrls); i++ {
		res := <-results
		bar.Add(1)
		if res.Success {
			fo.WriteString(fmt.Sprintf("%s - %s\n", res.Url, res.MatchedString))
		}
	}

}
