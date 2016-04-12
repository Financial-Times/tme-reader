package tme

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type Repository interface {
	GetTmeTermsFromIndex(int) ([]byte, error)
	GetTmeTermsInChunks(int, int) ([]byte, error)
	GetTmeTermById(string) ([]byte, error)
}

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type tmeRepository struct {
	httpClient   httpClient
	tmeBaseURL   string
	accessConfig tmeAccessConfig
	maxRecords   int
	slices       int
	taxonomyName string
}

type tmeAccessConfig struct {
	userName string
	password string
	token    string
}

func (t *tmeRepository) GetMaxRecords() int {
	return t.maxRecords
}

func (t *tmeRepository) GetTaxonomyName() string {
	return t.taxonomyName
}

func NewTmeRepository(client httpClient, tmeBaseURL string, userName string, password string, token string, maxRecords int, slices int, taxonomyName string) Repository {
	return &tmeRepository{httpClient: client, tmeBaseURL: tmeBaseURL, accessConfig: tmeAccessConfig{userName: userName, password: password, token: token}, maxRecords: maxRecords, slices: slices, taxonomyName: taxonomyName}
}

func (t *tmeRepository) GetTmeTermsFromIndex(startRecord int) ([]byte, error) {
	chunks := t.maxRecords / t.slices

	type dataChunkCollection struct {
		dataChunk []byte
		err       error
	}

	responseChannel := make(chan *dataChunkCollection, t.slices)
	go func() {
		var wg sync.WaitGroup
		wg.Add(t.slices)
		for i := 0; i < t.slices; i++ {
			startPosition := startRecord + i*chunks

			go func(startPosition int) {
				tmeTermsChunk, err := t.GetTmeTermsInChunks(startPosition, chunks)
				responseChannel <- &dataChunkCollection{dataChunk: tmeTermsChunk, err: err}
				wg.Done()
			}(startPosition)
		}
		wg.Wait()

		close(responseChannel)
	}()

	tmeTerms := []byte{}
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

func (t *tmeRepository) GetTmeTermsInChunks(startPosition int, maxRecords int) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/%s/terms?maximumRecords=%d&startRecord=%d", t.tmeBaseURL, t.taxonomyName, maxRecords, startPosition), nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("TME returned %d", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return contents, nil
}

func (t *tmeRepository) GetTmeTermById(rawId string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/authorityfiles/%s/terms/%s", t.tmeBaseURL, t.GetTaxonomyName(), rawId), nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Accept", "application/xml;charset=utf-8")
	req.SetBasicAuth(t.accessConfig.userName, t.accessConfig.password)
	req.Header.Add("X-Coco-Auth", fmt.Sprintf("%v", t.accessConfig.token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("TME returned %d HTTP status", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return contents, nil
}
