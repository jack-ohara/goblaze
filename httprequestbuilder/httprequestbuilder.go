package httprequestbuilder

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

func ExecuteGet(url string, headers map[string]string) []byte {
	req, _ := http.NewRequest("GET", url, nil)

	return executeRequest(req, headers)
}

func ExecutePost(url string, body []byte, headers map[string]string) []byte {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	return executeRequest(req, headers)
}

func executeRequest(request *http.Request, headers map[string]string) []byte {
	headers = addDefaultHeaders(headers)

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}

	resp, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return bodyBytes
}

func addDefaultHeaders(otherHeaders map[string]string) map[string]string {
	headers := map[string]string{
		"ContentType": "application/json",
		"Accept":      "application/json",
	}

	for k, v := range otherHeaders {
		headers[k] = v
	}

	return headers
}
