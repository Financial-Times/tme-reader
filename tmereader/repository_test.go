package tmereader

import (
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestGetTmeTermById(t *testing.T) {
	assert := assert.New(t)

	termFile := "sample_tme_term.xml"
	tmeTermsXML, err := os.Open(termFile)
	assert.Nil(err)

	body := ioutil.NopCloser(tmeTermsXML)

	tests := []struct {
		name         string
		repo         Repository
		expectedTerm term
		err          error
	}{{"Success",
		repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: body}, endpoint: "GL/terms/Nstein_GL_US_NY_Municipality_942968"}),
		term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"},
		nil,
	},
	}

	for _, test := range tests {
		actualTerm, err := test.repo.GetTmeTermById("Nstein_GL_US_NY_Municipality_942968")
		assert.Equal(test.expectedTerm, actualTerm, fmt.Sprintf("%s: Expected taxonomy incorrect ", test.name))
		assert.Equal(test.err, err)
	}
}

func TestGetTmeTermsInChunks(t *testing.T) {
	assert := assert.New(t)

	termFile := "sample_tme_terms.xml"
	tmeTermsXML, err := os.Open(termFile)
	assert.Nil(err)

	body := ioutil.NopCloser(tmeTermsXML)

	tests := []struct {
		name          string
		repo          Repository
		expectedTerms []term
		err           error
	}{{"Success",
		repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: body}, endpoint: "GL/terms?maximumRecords="}),
		[]term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
		nil,
	},
	}

	for _, test := range tests {
		actualTerms, err := test.repo.GetTmeTermsFromIndex(0)
		assert.Equal(len(test.expectedTerms), len(actualTerms), fmt.Sprintf("ExpectedTerms and ActualTerms vector size differ."))
		for _, expTerm := range test.expectedTerms {
			assert.Contains(actualTerms, expTerm, fmt.Sprintf("Actual taxonomy misses expected term %s", expTerm))
		}
		assert.Equal(test.err, err)
	}
}

func TestGetTmeTerms(t *testing.T) {
	assert := assert.New(t)

	termFile := "sample_tme_terms.xml"
	tmeTermsXML, err := os.Open(termFile)
	assert.Nil(err)

	body := ioutil.NopCloser(tmeTermsXML)

	tests := []struct {
		name          string
		repo          Repository
		expectedTerms []term
		err           error
	}{{"Success",
		repo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
			resp: http.Response{StatusCode: http.StatusOK, Body: body}, endpoint: "GL/terms?maximumRecords="}),
		[]term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
		nil,
	},
	}

	for _, test := range tests {
		actualTerms, err := test.repo.GetTmeTermsFromIndex(0)
		assert.Equal(len(test.expectedTerms), len(actualTerms), fmt.Sprintf("ExpectedTerms and ActualTerms vector size differ."))
		for _, expTerm := range test.expectedTerms {
			assert.Contains(actualTerms, expTerm, fmt.Sprintf("Actual taxonomy misses expected term %s", expTerm))
		}
		assert.Equal(test.err, err)
	}
}

type dummyTransformer struct {
}

type dummyModel struct {
	Terms []term `xml:"term"`
}

type term struct {
	CanonicalName string `xml:"name"`
	RawID         string `xml:"id"`
}

func (*dummyTransformer) UnMarshallTaxonomy(contents []byte) ([]interface{}, error) {
	taxonomy := dummyModel{}
	err := xml.Unmarshal(contents, &taxonomy)
	if err != nil {
		return nil, err
	}

	var interfaceSlice []interface{} = make([]interface{}, len(taxonomy.Terms))
	for i, d := range taxonomy.Terms {
		interfaceSlice[i] = d
	}
	return interfaceSlice, nil
}

func (*dummyTransformer) UnMarshallTerm(content []byte) (interface{}, error) {
	dummyTerm := term{}
	err := xml.Unmarshal(content, &dummyTerm)
	if err != nil {
		return term{}, err
	}
	return dummyTerm, nil
}

func repo(c dummyClient) Repository {
	return &tmeRepository{httpClient: &c, tmeBaseURL: c.tmeBaseURL, accessConfig: tmeAccessConfig{userName: "test", password: "test", token: "test"}, maxRecords: 100, slices: 1, taxonomyName: "GL", transformer: new(dummyTransformer)}
}

type dummyClient struct {
	assert     *assert.Assertions
	resp       http.Response
	err        error
	tmeBaseURL string
	endpoint   string
}

func (d *dummyClient) Do(req *http.Request) (resp *http.Response, err error) {
	d.assert.Contains(req.URL.String(), fmt.Sprintf("%s/rs/authorityfiles/%s", d.tmeBaseURL, d.endpoint), fmt.Sprintf("Expected url incorrect"))
	return &d.resp, d.err
}
