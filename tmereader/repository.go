package tmereader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type TmeSource interface {
	String() string
	PathSuffix() string
}

type AuthorityFiles struct{}
type KnowledgeBases struct{}

func (s *AuthorityFiles) String() string {
	return "authorityfiles"
}

func (s *AuthorityFiles) PathSuffix() string {
	return "terms"
}

func (s *KnowledgeBases) String() string {
	return "knowledgebases"
}

func (s *KnowledgeBases) PathSuffix() string {
	return "eng/categories"
}

type Repository interface {
	GetTmeTermsFromIndex(int) ([]interface{}, error)
	GetTmeTermById(string) (interface{}, error)
}

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type modelTransformer interface {
	UnMarshallTaxonomy(contents []byte) (tmeTerms []interface{}, error error)
	UnMarshallTerm(content []byte) (tmeTerm interface{}, error error)
}

type tmeRepository struct {
	httpClient   httpClient
	tmeBaseURL   string
	accessConfig TmeAccessConfig
	maxRecords   int
	slices       int
	taxonomyName string
	transformer  modelTransformer
	source       TmeSource
}

type TmeAccessConfig struct {
	userName string
	password string
	token    string
}

func NewTmeRepository(client httpClient, tmeBaseURL string, userName string, password string, token string, maxRecords int, slices int, taxonomyName string, source TmeSource, modelTransformer modelTransformer) Repository {
	return &tmeRepository{httpClient: client, tmeBaseURL: tmeBaseURL, accessConfig: TmeAccessConfig{userName: userName, password: password, token: token}, maxRecords: maxRecords, slices: slices, taxonomyName: taxonomyName, source: source, transformer: modelTransformer}
}

type TmeRepositoryConfig struct {
	client httpClient
	tmeBaseURL string
	userName string
	password string
	token string
	maxRecords int
	slices int
	taxonomyName string
	source TmeSource
	modelTransformer modelTransformer
}

func NewTimeRepositoryWithConfig(cfg TmeRepositoryConfig) Repository {
	return &tmeRepository{
		httpClient: cfg.client,
		tmeBaseURL: cfg.tmeBaseURL,
		accessConfig: TmeAccessConfig{userName: cfg.userName, password: cfg.password, token: cfg.token},
		maxRecords: cfg.maxRecords,
		slices: cfg.slices,
		taxonomyName: cfg.taxonomyName,
		source: cfg.source,
		transformer: cfg.modelTransformer,
	}
}

func (t *tmeRepository) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
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
			startPosition := startRecord + i*chunks

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
	url := fmt.Sprintf("%s/rs/%s/%s/%s?maximumRecords=%d&startRecord=%d", t.tmeBaseURL, t.source.String(), t.taxonomyName, t.source.PathSuffix(), maxRecords, startPosition)
	req, err := http.NewRequest("GET", url, nil)
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

	tmeTerms, err := t.transformer.UnMarshallTaxonomy(contents)
	if err != nil {
		return nil, err
	}
	return tmeTerms, nil
}

func (t *tmeRepository) GetTmeTermById(rawId string) (interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rs/%s/%s/%s/%s", t.tmeBaseURL, t.source.String(), t.taxonomyName, t.source.PathSuffix(), rawId), nil)
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
