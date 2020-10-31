package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/danielb42/whiteflag"
)

var (
	stringToLookFor = "ist bereits registriert"
	timeBegin       = time.Now()
	httpResp        *http.Response
	domainToCheck   string
	fails           int
)

func main() {
	whiteflag.Alias("d", "domain", "use the given domain for check order")
	domainToCheck = whiteflag.GetString("domain")

	run()
}

func run() {
	var err error
	log.SetOutput(os.Stderr)

	if httpResp != nil {
		httpResp.Body.Close() // nolint:errcheck
	}

	postString := fmt.Sprintf("lang=de&domain=%s&domainwhois_submit=Abfrage+starten", domainToCheck)
	postBody := strings.NewReader(postString)

	os.Setenv("HTTP_PROXY", "")
	os.Setenv("HTTPS_PROXY", "")
	os.Setenv("http_proxy", "")
	os.Setenv("https_proxy", "")

	httpReq, err := http.NewRequest("POST", "https://www.denic.de/webwhois/", postBody)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpResp, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	webwhoisResponseTime := time.Since(timeBegin).Milliseconds()

	bodyBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		printFailMetricsAndExit(err.Error())
	}

	if strings.Contains(string(bodyBytes), stringToLookFor) {
		log.Printf("OK: webwhois output contains '%s'\n\n", stringToLookFor)
		fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d %d\n",
			"webwhois",
			"registered", 1,
			"duration", webwhoisResponseTime,
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

	fmt.Printf("extmon,service=%s %s=%d,%s=%d,%s=%d %d\n",
		"webwhois",
		"registered", 0,
		"duration", 0,
		"responsecode", statusCode,
		timeBegin.Unix())

	os.Exit(2)
}
