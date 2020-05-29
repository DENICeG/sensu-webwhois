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
	domainToCheck   string
	stringToLookFor string = "ist bereits registriert."
)

func setAliasesViaWhiteflag() {
	whiteflag.Alias("dom", "domain", "use the given domain for check order")
}

func main() {
	log.SetOutput(os.Stderr)

	// Parse commandline parameters

	setAliasesViaWhiteflag()
	whiteflag.ParseCommandLine()
	domainToCheck = whiteflag.GetString("domain")

	postString := fmt.Sprintf("lang=de&domain=%s&domainwhois_submit=Abfrage+starten", domainToCheck)
	postbody := strings.NewReader(postString)

	// body := strings.NewReader(`lang=de&domain=denic.de&domainwhois_submit=Abfrage+starten`)

	timeBegin := time.Now()
	req, err := http.NewRequest("POST", "https://www.denic.de/webwhois/", postbody)
	if err != nil {
		log.Println(err)
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
		os.Exit(2)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	timeEnd := time.Since(timeBegin)
	if err != nil {
		log.Println(err)
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
		os.Exit(2)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
			os.Exit(2)
		}
		bodyString := string(bodyBytes)

		if strings.Contains(bodyString, stringToLookFor) {
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 1, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", timeEnd.Milliseconds(), timeBegin.Unix())
			os.Exit(0)
		}

	}

	fmt.Printf("ERROR: HTTP Code was not 200")
	fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
	fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", timeEnd.Milliseconds(), timeBegin.Unix())
	os.Exit(2)
}
