package mian

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	myInstanceID   string
	routerIPs      []string
	cfPingInstance string // APP-GUID:INSTNACE-INDEX-NUMBER
	appDomain      string
	sleepInterval  int64
)

type ResponseBody struct {
	CFIndex string `json:"cf-instance-index"`
	Status  string `json:"status"`
}

func init() {
	myInstanceID = os.Getenv("CF_INSTANCE_INDEX")
	routers := os.Getenv("GOROUTER_LIST")
	cfPingInstance = os.Getenv("CF_PING_INSTANCE")
	appDomain = os.Getenv("CF_APP_DOMAIN")
	si := os.Getenv("PING_SLEEP_INTERVAL_SECONDS")

	if myInstanceID == "" {
		panic("env var CF_INSTANCE_INDEX not set")
	}

	routerIPs = make([]string, 0)
	if routers != "" {
		routerIPs = strings.Split(routers, ":")
	}

	if cfPingInstance == "" {
		panic("env var CF_PING_INSTANCE not set")
	}

	if si != "" {
		n, err := strconv.Atoi(si)
		if err != nil {
			panic(fmt.Sprintf("sleep interval invalid: %s", err))
		}
		sleepInterval = int64(n)
	} else {
		sleepInterval = 2
	}
}

func pingInstances() {

	pi := strings.Split(cfPingInstance, ":")
	if len(pi) != 2 {
		panic("CF_PING_INSTANCE is malformed")
	}

	if pi[1] == myInstanceID {
		fmt.Println("ping instnace index matches my index.  Shutting down pings")
		return
	}

	for {
		time.Sleep(time.Duration(sleepInterval * int64(time.Second)))

		for _, r := range routerIPs {
			time.Sleep(time.Duration(time.Millisecond * 100)) // sleep between gorouter pings
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}

			req, err := http.NewRequest(http.MethodGet, r, nil)
			if err != nil {
				fmt.Printf("client: could not create request: %s\n", err)
				continue
			}
			req.Header.Add("X-Cf-App-Instance", cfPingInstance)
			req.Host = appDomain
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("failed to send request via router %s to app %s at instance %s: %s\n", r, appDomain, cfPingInstance, err)
				continue
			}
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("could not read response body: %s\n", err)
				continue
			}
			res.Body.Close()

			if res.StatusCode != 200 {
				fmt.Printf("bad response code %d\n%s\n", res.StatusCode, resBody)
				continue
			}
			fmt.Printf("successful response: %s\n", resBody)
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	res := ResponseBody{myInstanceID, "i am good"}
	b, err := json.Marshal(&res)
	if err != nil {
		fmt.Printf("Failed to marhsal response: %s\n", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func main() {

	go pingInstances()
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
