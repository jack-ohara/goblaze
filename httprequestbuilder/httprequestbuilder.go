package httprequestbuilder

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

type HttpResponse struct {
	StatusCode  int
	BodyContent []byte
	Headers     map[string][]string
}

func ExecuteGet(url string, headers map[string]string) HttpResponse {
	req, _ := http.NewRequest("GET", url, nil)

	req.Close = true

	return executeRequest(req, headers)
}

func ExecutePost(url string, body []byte, headers map[string]string) HttpResponse {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Close = true

	return executeRequest(req, headers)
}

func executeRequest(request *http.Request, headers map[string]string) HttpResponse {
	headers = addDefaultHeaders(headers)

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}

	resp, err := client.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	httpResponse := HttpResponse{
		BodyContent: bodyBytes,
		Headers:     resp.Header,
		StatusCode:  resp.StatusCode,
	}

	return httpResponse
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
