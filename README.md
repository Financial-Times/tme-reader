# tme-reader

[![Circle CI](https://circleci.com/gh/Financial-Times/tme-reader/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/tme-reader/tree/master)

Retrieves General Terms from TME as a list of interfaces, letting the main application to decide how an output will look like.

The service exposes endpoints for getting all the terms and for getting a term by a tmeID.


# Usage
`go get github.com/Financial-Times/tme-reader/tmereader`

Available methods:

* GetTmeTermsFromIndex(index int) (tmeTerms []interface{}, error) - returns a set of terms, having a maximum of `maxRecord` elements starting from the provided index 	
* GetTmeTermById(tmeID string) (tmeTerm interface{}, error) - returns the term details, obtained by the tme term identifier

# In order to run:

Create a new repository:

`NewTmeRepository(client httpClient, tmeBaseURL string, userName string, password string, token string, maxRecords int, slices int, taxonomyName string, modelTransformer modelTransformer)`

The modelTransformer should implement the following methods, according to his own model type:

* UnMarshallTaxonomy(contents []byte) (tmeTerms []interface{}, error) - loading xml data into list of terms ([]tmeTerms)
* UnMarshallTerm(content []byte) (tmeTerm interface{}, error) - loading xml data into a tmeTerm model

