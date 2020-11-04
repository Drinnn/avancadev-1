package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/hashicorp/go-retryablehttp"
)

type Result struct {
	Status string
}

func main() {
	http.HandleFunc("/", home)
	http.ListenAndServe(":9091", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	coupon := r.PostFormValue("coupon")
	ccNumber := r.PostFormValue("ccNumber")

	resultCoupon := makeHTTPCall("http://localhost:9092", coupon)

	result := Result{Status: "declined"}

	if ccNumber == "1" {
		result.Status = "approved"
	}

	if resultCoupon.Status == "invalid" {
		result.Status = "invalid coupon"
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Fatal("Error processing JSON.")
	}

	fmt.Fprintf(w, string(jsonData))
}

func process(w http.ResponseWriter, r *http.Request) {
	result := makeHTTPCall("http://localhost:9091", r.FormValue("coupon"))

	t := template.Must(template.ParseFiles("templates/home.html"))
	t.Execute(w, result)
}

func makeHTTPCall(urlMicroservice string, coupon string) Result {
	values := url.Values{}
	values.Add("coupon", coupon)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	res, err := retryClient.PostForm(urlMicroservice, values)
	if err != nil {
		result := Result{Status: "Servivor fora do ar!"}
		return result
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Error processing result.")
	}

	result := Result{}

	json.Unmarshal(data, &result)

	return result
}
