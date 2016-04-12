package tme

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestGetTmeTermsTaxonomy(t *testing.T) {
	termFile := "sample_tme_terms.xml"

	assert := assert.New(t)
	tmeTermsXML, err := os.Open(termFile)
	body := ioutil.NopCloser(tmeTermsXML)
	data, err := ioutil.ReadFile(termFile)

	log.Printf("%v\n", err)

	tests := []struct {
		name string
		repo Repository
		tax  []byte
		err  error
	}{
		{"Success", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: body}}), data, nil},
		{"Error", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: body}, err: errors.New("Some error")}),
			[]byte{}, errors.New("Some error")},
		{"Non 200 from structure service", repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusBadRequest, Body: body}}),
			[]byte{}, errors.New("TME returned 400")},
	}

	for _, test := range tests {
		expectedTax, err := test.repo.GetTmeTermsFromIndex(0)
		assert.Equal(test.tax, expectedTax, fmt.Sprintf("%s: Expected taxonomy incorrect", test.name))
		assert.Equal(test.err, err)
	}

}

func repo(c dummyClient) Repository {
	return &tmeRepository{httpClient: &c, tmeBaseURL: c.tmeBaseURL, accessConfig: tmeAccessConfig{userName: "test", password: "test", token: "test"}, maxRecords: 2, slices: 2, taxonomyName: "GL"}
}

type dummyClient struct {
	assert     *assert.Assertions
	resp       http.Response
	err        error
	tmeBaseURL string
}

func (d *dummyClient) Do(req *http.Request) (resp *http.Response, err error) {
	d.assert.Contains(req.URL.String(), fmt.Sprintf("%s/rs/authorityfiles/GL/terms?maximumRecords=", d.tmeBaseURL), fmt.Sprintf("Expected url incorrect"))
	return &d.resp, d.err
}
