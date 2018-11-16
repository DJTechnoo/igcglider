package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"
)

func TestTrackParse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(inputHandler))
	defer ts.Close()

	testURL := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	resp, err := http.PostForm(ts.URL, url.Values{"url": {testURL}})
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("File couldn't load")
	}

	resp, err = http.PostForm(ts.URL, url.Values{"url": {" "}})
	if err != nil {
		t.Errorf("Error creating the POST request, %s", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Error("Expected bad request")
	}

}
