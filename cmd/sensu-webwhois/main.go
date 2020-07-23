package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	whiteflag "github.com/danielb42/whiteflag" // MIT
)

var (
	stringToLookFor string = "ist bereits registriert."
)

func main() {
	log.SetOutput(os.Stderr)

	whiteflag.Alias("d", "domain", "use the given domain for check order")
	whiteflag.ParseCommandLine()
	domainToCheck := whiteflag.GetString("domain")

	postString := fmt.Sprintf("lang=de&domain=%s&domainwhois_submit=Abfrage+starten", domainToCheck)
	postbody := strings.NewReader(postString)

	timeBegin := time.Now()

	req, err := http.NewRequest("POST", "https://www.denic.de/webwhois/", postbody)
	if err != nil {
		log.Printf("ERROR: %s\n\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", 0, timeBegin.Unix())
		os.Exit(2)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERROR: %s\n\n", err.Error())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", 0, timeBegin.Unix())
		os.Exit(2)
	}
	defer resp.Body.Close()

	webwhoisResponseTime := time.Since(timeBegin).Milliseconds()

	if resp.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: %s\n\n", err.Error())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", 200, timeBegin.Unix())
			os.Exit(2)
		}
		bodyString := string(bodyBytes)

		if strings.Contains(bodyString, stringToLookFor) {
			log.Printf("OK: webwhois output contains '%s'\n\n", stringToLookFor)
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 1, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", 200, timeBegin.Unix())
			os.Exit(0)
		} else {
			log.Printf("ERROR: webwhois output did not contain '%s'\n\n", stringToLookFor)
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", 200, timeBegin.Unix())
			os.Exit(2)
		}
	} else {
		log.Printf("ERROR: HTTP status code was not 200\n\n")
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.responsecode", resp.StatusCode, timeBegin.Unix())

		os.Exit(2)
	}
}
