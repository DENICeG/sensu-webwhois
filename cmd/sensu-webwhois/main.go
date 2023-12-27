package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

const (
	stringToLookFor = "ist bereits registriert"
)

var (
	domainToCheck    = pflag.StringP("domain", "d", "denic.de", "use the given domain for check order")
	webWhoisURL      = pflag.StringP("address", "a", "https://webwhois.denic.de/", "full webwhois url")
	webWhoisInsecure = pflag.Bool("insecure", false, "disable tls certificate check")
	retries          = pflag.Int("retries", 3, "number of retries when the request fails")
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("UTC ")
	log.SetFlags(log.Ltime | log.Lmsgprefix | log.LUTC)

	pflag.Parse()

	result := Execute()

	fmt.Println()
	var available int
	if result.Success {
		available = 1
	}
	// influx-db format
	fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d,%s=%d %d\n",
		"webwhois",
		"available", available,
		"registered", available,
		"duration", result.Duration.Milliseconds(),
		"order", result.Duration.Milliseconds(),
		"responsecode", result.ResponseCode,
		result.StartTime.Unix())

	if !result.Success {
		os.Exit(2)
		return
	}
}

type Result struct {
	StartTime    time.Time
	Duration     time.Duration
	Success      bool
	ResponseCode int
}

func Execute() (result Result) {
	os.Setenv("HTTP_PROXY", "")
	os.Setenv("HTTPS_PROXY", "")
	os.Setenv("http_proxy", "")
	os.Setenv("https_proxy", "")

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *webWhoisInsecure,
			},
		},
	}

	result.StartTime = time.Now()
	defer func() {
		result.Duration = time.Since(result.StartTime)
	}()
	for i := 0; i <= *retries; i++ {
		result.ResponseCode = 0

		req, err := http.NewRequest(http.MethodGet, *webWhoisURL, nil)
		if err != nil {
			// this error will not resolve with retries
			log.Printf("ERROR: %s\n\n", err.Error())
			break
		}

		domainQuery := req.URL.Query()
		domainQuery.Add("lang", "de")
		domainQuery.Add("query", *domainToCheck)

		req.URL.RawQuery = domainQuery.Encode()

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
			continue
		}

		result.ResponseCode = resp.StatusCode

		var outBody string
		if resp.Body != nil {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("ERROR: %s\n", err.Error())
				continue
			}
			outBody = string(bodyBytes)
		}

		if strings.Contains(outBody, stringToLookFor) {
			result.Success = true
			log.Printf("OK: webwhois output contains '%s'\n", stringToLookFor)
			break
		} else {
			log.Printf("ERROR: webwhois output did not contain '%s'\n", stringToLookFor)
			continue
		}
	}
	return
}
