package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	kh "github.com/Comcast/kuberhealthy/v2/pkg/checks/external/checkclient"
	log "github.com/sirupsen/logrus"
)

var (
	// Environment Variables fetched from spec file
	checkURL = os.Getenv("CHECK_URL")
	count    = os.Getenv("COUNT")
	seconds  = os.Getenv("SECONDS")
	passing  = os.Getenv("PASSING_PERCENT")
)

func init() {
	// Check that the URL environment variable is valid.
	if len(checkURL) == 0 {
		err := fmt.Errorf("empty CHECK_URL specified. Please update your CHECK_URL environment variable")
		ReportFailureAndPanic(err)
	}

	// Check that the COUNT environment variable is valid.
	if len(count) == 0 {
		count = "0"
	}

	// Check that the SECONDS environment variable is valid.
	if len(seconds) == 0 {
		seconds = "0"
	}

	// If the URL does not begin with HTTP, exit.
	if !strings.HasPrefix(checkURL, "http") {
		err := fmt.Errorf("given URL does not declare a supported protocol. (http | https)")
		ReportFailureAndPanic(err)
	}
}

func main() {

	countInt, err := strconv.Atoi(count)
	if err != nil {
		err = fmt.Errorf("Error converting COUNT to int: " + err.Error())
		ReportFailureAndPanic(err)
	}

	secondInt, err := strconv.Atoi(seconds)
	if err != nil {
		err = fmt.Errorf("Error converting SECONDS to int: " + err.Error())
		ReportFailureAndPanic(err)
	}

	passingInt, err := strconv.Atoi(passing)
	if err != nil {
		err = fmt.Errorf("Error converting PASSING_PERCENT to int: " + err.Error())
		ReportFailureAndPanic(err)
	}

	// if the passing count is empty, then default to 1
	if passingInt == 0 {
		passingInt = 1
	}
	passingPercentage := passingInt / 100

	// sets the passing score to compare it against checks that have been ran
	passingScore := passingPercentage * countInt
	passInt := passingScore
	log.Infoln("Looking for at least", passingInt, "percent of", countInt, "checks to pass")

	// init counters for checks
	log.Infoln("Beginning check.")
	checksRan := 0
	checksPassed := 0
	checksFailed := 0

	ticker := time.NewTicker(time.Duration(secondInt))
	defer ticker.Stop()

	// This for loop makes a http GET request to a known internet address, address can be changed in deployment spec yaml
	// and returns a http status every second.
	for checksRan < countInt {
		<-ticker.C
		r, err := http.Get(checkURL)
		checksRan++

		if err != nil {
			checksFailed++
			log.Errorln("Failed to reach URL: ", checkURL)
			continue
		}

		if r.StatusCode != http.StatusOK {
			log.Errorln("Got a", r.StatusCode, "with a", http.MethodGet, "to", checkURL)
			checksFailed++
			continue
		}
		log.Errorln("Got a", r.StatusCode, "with a", http.MethodGet, "to", checkURL)
		checksPassed++
	}

	// Displays the results of 10 URL requests
	log.Infoln(checksRan, "checks ran")
	log.Infoln(checksPassed, "checks passed")
	log.Infoln(checksFailed, "checks failed")

	// Check to see if the number of requests passed at passingPercent and reports to Kuberhealthy accordingly
	if checksPassed < passInt {
		reportErr := fmt.Errorf("unable to retrieve a valid response from " + checkURL + "check failed " + strconv.Itoa(checksFailed) + " out of " + strconv.Itoa(checksRan) + " attempts")
		ReportFailureAndPanic(reportErr)
	}

	err = kh.ReportSuccess()
	if err != nil {
		log.Fatalln("error when reporting to kuberhealthy:", err.Error())
	}
	log.Infoln("Successfully reported to Kuberhealthy")
}

// ReportFailureAndPanic logs and reports an error to kuberhealthy and then fatals the program
func ReportFailureAndPanic(err error) {
	log.Errorln(err)
	err2 := kh.ReportFailure([]string{err.Error()})
	if err2 != nil {
		log.Fatalln("error when reporting to kuberhealthy:", err.Error())
	}
	os.Exit(0)
}
