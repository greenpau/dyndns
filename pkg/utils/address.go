package utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const checkipURL string = "https://checkip.amazonaws.com/"

// NewBrowser returns HTTP client.
func NewBrowser() (*http.Client, error) {
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookier jar error: %s", err)
	}

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 3 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 3 * time.Second,
	}

	client := &http.Client{
		Jar:       cj,
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	return client, nil
}

// GetPublicAddress returns public IP address of the host
// where this function is running on.
func GetPublicAddress(version int) (string, error) {
	if version != 4 {
		return "", fmt.Errorf("only ip version 4 is supported")
	}

	browser, err := NewBrowser()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", checkipURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating http get: %s", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	resp, err := browser.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %s", err)
	}
	responseBody := string(body[:])
	responseBody = strings.TrimSpace(responseBody)
	address := net.ParseIP(responseBody)
	if address == nil {
		return "", fmt.Errorf("error parsing ip address from %s", responseBody)
	}
	if address.String() != responseBody {
		return "", fmt.Errorf("error parsing ip address: %s (received) vs. %s (parsed)", responseBody, address.String())
	}

	return address.String(), nil
}
