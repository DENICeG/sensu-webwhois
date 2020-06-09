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

	req, err := http.NewRequest("POST", "https://www.denic.de/webwhois/", postbody)
	if err != nil {
		log.Println(err)
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, time.Now().Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, time.Now().Unix())
		os.Exit(3)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	timeBegin := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", 0, timeBegin.Unix())
		os.Exit(2)
	}
	defer resp.Body.Close()

	webwhoisResponseTime := time.Since(timeBegin).Milliseconds()

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
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
			os.Exit(0)
		} else {
			log.Printf("error: webwhois output did not contain '%s'", stringToLookFor)
			fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
			fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
			os.Exit(2)
		}
	} else {
		log.Println("error: HTTP status code was not 200")
		fmt.Printf("%s %d %d\n", "sensu.webwhois.registered", 0, timeBegin.Unix())
		fmt.Printf("%s %d %d\n", "sensu.webwhois.duration", webwhoisResponseTime, timeBegin.Unix())
		os.Exit(2)
	}
}
