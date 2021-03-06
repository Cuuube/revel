// Copyright (c) 2012-2016 The Revel Framework Authors, All rights reserved.
// Revel Framework source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package revel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

// Params: Testing Multipart forms

const (
	MultipartBoundary = "A"
	MultipartFormData = `--A
Content-Disposition: form-data; name="text1"

data1
--A
Content-Disposition: form-data; name="text2"

data2
--A
Content-Disposition: form-data; name="text2"

data3
--A
Content-Disposition: form-data; name="file1"; filename="test.txt"
Content-Type: text/plain

content1
--A
Content-Disposition: form-data; name="file2[]"; filename="test.txt"
Content-Type: text/plain

content2
--A
Content-Disposition: form-data; name="file2[]"; filename="favicon.ico"
Content-Type: image/x-icon

xyz
--A
Content-Disposition: form-data; name="file3[0]"; filename="test.txt"
Content-Type: text/plain

content3
--A
Content-Disposition: form-data; name="file3[1]"; filename="favicon.ico"
Content-Type: image/x-icon

zzz
--A--
`
)

// The values represented by the form data.
type fh struct {
	filename string
	content  []byte
}

var (
	expectedValues = map[string][]string{
		"text1": {"data1"},
		"text2": {"data2", "data3"},
	}
	expectedFiles = map[string][]fh{
		"file1":    {fh{"test.txt", []byte("content1")}},
		"file2[]":  {fh{"test.txt", []byte("content2")}, fh{"favicon.ico", []byte("xyz")}},
		"file3[0]": {fh{"test.txt", []byte("content3")}},
		"file3[1]": {fh{"favicon.ico", []byte("zzz")}},
	}
)

func getMultipartRequest() *http.Request {
	req, _ := http.NewRequest("POST", "http://localhost/path",
		bytes.NewBufferString(MultipartFormData))
	req.Header.Set(
		"Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", MultipartBoundary))
	req.Header.Set(
		"Content-Length", fmt.Sprintf("%d", len(MultipartFormData)))
	return req
}

func BenchmarkParams(b *testing.B) {
	c := NewTestController(nil, showRequest)
	c.Params = &Params{}

	for i := 0; i < b.N; i++ {
		ParamsFilter(c, NilChain)
	}
}

func TestMultipartForm(t *testing.T) {
	c := NewTestController(nil, getMultipartRequest())
	c.Params = &Params{}

	ParamsFilter(c, NilChain)

	if !reflect.DeepEqual(expectedValues, map[string][]string(c.Params.Values)) {
		t.Errorf("Param values: (expected) %v != %v (actual)",
			expectedValues, map[string][]string(c.Params.Values))
	}

	actualFiles := make(map[string][]fh)
	for key, fileHeaders := range c.Params.Files {
		for _, fileHeader := range fileHeaders {
			file, _ := fileHeader.Open()
			content, _ := ioutil.ReadAll(file)
			actualFiles[key] = append(actualFiles[key], fh{fileHeader.Filename, content})
		}
	}

	if !reflect.DeepEqual(expectedFiles, actualFiles) {
		t.Errorf("Param files: (expected) %v != %v (actual)", expectedFiles, actualFiles)
	}
}

func TestBind(t *testing.T) {
	params := Params{
		Values: url.Values{
			"x": {"5"},
		},
	}
	var x int
	params.Bind(&x, "x")
	if x != 5 {
		t.Errorf("Failed to bind x.  Value: %d", x)
	}
}

func TestResolveAcceptLanguage(t *testing.T) {
	request := buildHTTPRequestWithAcceptLanguage("")
	if result := ResolveAcceptLanguage(request); result != nil {
		t.Errorf("Expected Accept-Language to resolve to an empty string but it was '%s'", result)
	}

	request = buildHTTPRequestWithAcceptLanguage("en-GB,en;q=0.8,nl;q=0.6")
	if result := ResolveAcceptLanguage(request); len(result) != 3 {
		t.Errorf("Unexpected Accept-Language values length of %d (expected %d)", len(result), 3)
	} else {
		if result[0].Language != "en-GB" {
			t.Errorf("Expected '%s' to be most qualified but instead it's '%s'", "en-GB", result[0].Language)
		}
		if result[1].Language != "en" {
			t.Errorf("Expected '%s' to be most qualified but instead it's '%s'", "en", result[1].Language)
		}
		if result[2].Language != "nl" {
			t.Errorf("Expected '%s' to be most qualified but instead it's '%s'", "nl", result[2].Language)
		}
	}

	request = buildHTTPRequestWithAcceptLanguage("en;q=0.8,nl;q=0.6,en-AU;q=malformed")
	if result := ResolveAcceptLanguage(request); len(result) != 3 {
		t.Errorf("Unexpected Accept-Language values length of %d (expected %d)", len(result), 3)
	} else {
		if result[0].Language != "en-AU" {
			t.Errorf("Expected '%s' to be most qualified but instead it's '%s'", "en-AU", result[0].Language)
		}
	}
}

func BenchmarkResolveAcceptLanguage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		request := buildHTTPRequestWithAcceptLanguage("en-GB,en;q=0.8,nl;q=0.6,fr;q=0.5,de-DE;q=0.4,no-NO;q=0.4,ru;q=0.2")
		ResolveAcceptLanguage(request)
	}
}

func buildHTTPRequestWithAcceptLanguage(acceptLanguage string) *Request {
	request, _ := http.NewRequest("POST", "http://localhost/path", nil)
	request.Header.Set("Accept-Language", acceptLanguage)
	c := NewTestController(nil, request)

	return c.Request
}
