package tmereader

import (
	"encoding/xml"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestGetTmeTermById(t *testing.T) {
	assert := assert.New(t)

	bodyForAuthorityFiles := getFileReader(assert, "sample_tme_term.xml")
	bodyForKnowledgeBases := getFileReader(assert, "sample_tme_term.xml")

	tests := []struct {
		name         string
		repo         Repository
		expectedTerm term
		err          error
	}{
		{"Success",
			authorityFilesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForAuthorityFiles}, source: &AuthorityFiles{}, endpoint: "GL/terms/Nstein_GL_US_NY_Municipality_942968"}),
			term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"},
			nil,
		},
		{"Success",
			knowledgeBasesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForKnowledgeBases}, source: &KnowledgeBases{}, endpoint: "GL/eng/categories/Nstein_GL_US_NY_Municipality_942968"}),
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

	bodyForAuthorityFiles := getFileReader(assert, "sample_tme_terms.xml")
	bodyForKnowledgeBases := getFileReader(assert, "sample_tme_terms.xml")

	tests := []struct {
		name          string
		repo          Repository
		expectedTerms []term
		err           error
	}{
		{"Success",
			authorityFilesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForAuthorityFiles}, source: &AuthorityFiles{}, endpoint: "GL/terms?maximumRecords="}),
			[]term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			nil,
		},
		{"Success",
			knowledgeBasesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForKnowledgeBases}, source: &KnowledgeBases{}, endpoint: "GL/eng/categories?maximumRecords="}),
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

	bodyForAuthorityFiles := getFileReader(assert, "sample_tme_terms.xml")
	bodyForKnowledgeBases := getFileReader(assert, "sample_tme_terms.xml")

	tests := []struct {
		name          string
		repo          Repository
		expectedTerms []term
		err           error
	}{
		{"Success",
			authorityFilesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForAuthorityFiles}, source: &AuthorityFiles{}, endpoint: "GL/terms?maximumRecords="}),
			[]term{term{CanonicalName: "Banksville, New York", RawID: "Nstein_GL_US_NY_Municipality_942968"}},
			nil,
		},
		{"Success",
			knowledgeBasesRepo(dummyClient{assert: assert, tmeBaseURL: "https://test-url.com:40001",
				resp: http.Response{StatusCode: http.StatusOK, Body: bodyForKnowledgeBases}, source: &KnowledgeBases{}, endpoint: "GL/eng/categories?maximumRecords="}),
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

func authorityFilesRepo(c dummyClient) Repository {
	return NewTimeRepositoryWithConfig(TmeRepositoryConfig{
		client: &c,
		tmeBaseURL: c.tmeBaseURL,
		userName: "test",
		password: "test",
		token: "test",
		maxRecords: 100,
		slices: 1,
		taxonomyName: "GL",
		source: &AuthorityFiles{},
		modelTransformer: new(dummyTransformer),
	})
}

func knowledgeBasesRepo(c dummyClient) Repository {
	return NewTimeRepositoryWithConfig(TmeRepositoryConfig{
		client: &c,
		tmeBaseURL: c.tmeBaseURL,
		userName: "test",
		password: "test",
		token: "test",
		maxRecords: 100,
		slices: 1,
		taxonomyName: "GL",
		source: &KnowledgeBases{},
		modelTransformer: new(dummyTransformer),
	})
}

func getFileReader(assert *assert.Assertions, name string) io.ReadCloser {
	file, err := os.Open(name)
	assert.Nil(err)

	return ioutil.NopCloser(file)
}

type dummyClient struct {
	assert     *assert.Assertions
	resp       http.Response
	err        error
	tmeBaseURL string
	source     TmeSource
	endpoint   string
}

func (d *dummyClient) Do(req *http.Request) (resp *http.Response, err error) {
	d.assert.Contains(req.URL.String(), fmt.Sprintf("%s/rs/%s/%s", d.tmeBaseURL, d.source.String(), d.endpoint), fmt.Sprintf("Expected url incorrect"))
	return &d.resp, d.err
}
