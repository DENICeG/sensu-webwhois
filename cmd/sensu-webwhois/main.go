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

	"github.com/danielb42/whiteflag"
)

var (
	stringToLookFor  = "ist bereits registriert"
	timeBegin        = time.Now()
	httpResp         *http.Response
	domainToCheck    string
	webWhoisURL      string
	webWhoisInsecure bool
	fails            int
)

func main() {
	whiteflag.Alias("d", "domain", "use the given domain for check order")
	whiteflag.Alias("a", "address", "full webwhois url")
	domainToCheck = whiteflag.GetString("domain")
	if whiteflag.FlagPresent("address") {
		webWhoisURL = whiteflag.GetString("address")
	} else {
		webWhoisURL = "https://www.denic.de/webwhois/"
	}
	webWhoisInsecure = whiteflag.FlagPresent("insecure")

	run()
}

func run() {
	var err error
	log.SetOutput(os.Stderr)
	log.SetPrefix("UTC ")
	log.SetFlags(log.Ltime | log.Lmsgprefix | log.LUTC)

	if httpResp != nil {
		httpResp.Body.Close() // nolint:errcheck
	}

	os.Setenv("HTTP_PROXY", "")
	os.Setenv("HTTPS_PROXY", "")
	os.Setenv("http_proxy", "")
	os.Setenv("https_proxy", "")

	httpReq, err := http.NewRequest(http.MethodGet, webWhoisURL, nil)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	domainQuery := httpReq.URL.Query()
	domainQuery.Add("lang", "de")
	domainQuery.Add("query", domainToCheck)
	domainQuery.Add("domain", domainToCheck)
	domainQuery.Add("domainwhois_submit", "Abfrage+starten")

	httpReq.URL.RawQuery = domainQuery.Encode()

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: webWhoisInsecure,
			},
		},
	}
	httpResp, err = httpClient.Do(httpReq)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	webwhoisResponseTime := time.Since(timeBegin).Milliseconds()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	if strings.Contains(string(bodyBytes), stringToLookFor) {
		log.Printf("OK: webwhois output contains '%s'\n\n", stringToLookFor)
		fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d,%s=%d %d\n",
			"webwhois",
			"available", 1,
			"registered", 1,
			"duration", webwhoisResponseTime,
			"order", webwhoisResponseTime,
			"responsecode", httpResp.StatusCode,
			timeBegin.Unix())
	} else {
		printFailMetricsAndExit("webwhois output did not contain", "'"+stringToLookFor+"'")
	}

	httpResp.Body.Close()
	os.Exit(0)
}

func printFailMetricsAndExit(errors ...string) {

	if fails < 3 {
		fails++
		run()
	}

	var statusCode int

	if httpResp != nil {
		statusCode = httpResp.StatusCode
		httpResp.Body.Close() // nolint:errcheck
	}

	errStr := "ERROR:"

	for _, err := range errors {
		errStr += " " + err
	}

	log.Printf("%s\n\n", errStr)

	fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d,%s=%d,%s=%d %d\n",
		"webwhois",
		"available", 0,
		"registered", 0,
		"duration", 0,
		"order", 0,
		"responsecode", statusCode,
		timeBegin.Unix())

	os.Exit(2)
}
