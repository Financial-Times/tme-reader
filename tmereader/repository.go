package tmereader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type Repository interface {
	GetTmeTerms() ([]interface{}, error)
	GetTmeTermById(string) (interface{}, error)
}

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type modelTransformer interface {
	UnMarshallTaxonomy([]byte) (interface{}, error)
	UnMarshallTerm([]byte) (interface{}, error)
	GetTermsFromTaxonomy(interface{}) []interface{}
}

type tmeRepository struct {
	httpClient   httpClient
	tmeBaseURL   string
	accessConfig tmeAccessConfig
	maxRecords   int
	slices       int
	taxonomyName string
	transformer  modelTransformer
}

type tmeAccessConfig struct {
	userName string
	password string
	token    string
}

func NewTmeRepository(client httpClient, tmeBaseURL string, userName string, password string, token string, maxRecords int, slices int, taxonomyName string, modelTransformer modelTransformer) Repository {
	return &tmeRepository{httpClient: client, tmeBaseURL: tmeBaseURL, accessConfig: tmeAccessConfig{userName: userName, password: password, token: token}, maxRecords: maxRecords, slices: slices, taxonomyName: taxonomyName, transformer: modelTransformer}
}

func (t *tmeRepository) GetTmeTerms() ([]interface{}, error) {
	chunks := t.maxRecords / t.slices

	type dataChunkCollection struct {
		dataChunk []interface{}
		err       error
	}

	responseChannel := make(chan *dataChunkCollection, t.slices)
	go func() {
		var wg sync.WaitGroup
		wg.Add(t.slices)
		for i := 0; i < t.slices; i++ {
			startPosition := i * chunks

			go func(startPosition int) {
				tmeTermsChunk, err := t.getTmeTermsInChunks(startPosition, chunks)
				responseChannel <- &dataChunkCollection{dataChunk: tmeTermsChunk, err: err}
				wg.Done()
			}(startPosition)
		}
		wg.Wait()

		close(responseChannel)
	}()

	var tmeTerms []interface{}
	var err error = nil
	for response := range responseChannel {
		if response.err != nil {
			err = response.err
		} else {
			tmeTerms = append(tmeTerms, response.dataChunk...)
		}
	}
	return tmeTerms, err
}

func (t *tmeRepository) getTmeTermsInChunks(startPosition int, maxRecords int) ([]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/%s/terms?maximumRecords=%d&startRecord=%d", t.tmeBaseURL, t.taxonomyName, maxRecords, startPosition), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TME returned %d", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	taxonomy, err := t.transformer.UnMarshallTaxonomy(contents)
	if err != nil {
		return nil, err
	}
	return t.transformer.GetTermsFromTaxonomy(taxonomy), nil
}

func (t *tmeRepository) GetTmeTermById(rawId string) (interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/%s/terms/%s", t.tmeBaseURL, t.taxonomyName, rawId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TME returned %d HTTP status", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return t.transformer.UnMarshallTerm(contents)
}
