tme-reader
==========

[![Circle CI](https://circleci.com/gh/Financial-Times/tme-reader/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/tme-reader/tree/master)

Retrieves General Terms from TME as a list of interfaces, letting the main application to decide how an output will look like.

The service exposes endpoints for getting all the terms and for getting a term by a tmeID.


Usage
-----

    go get github.com/Financial-Times/tme-reader/tmereader

Available methods:

* `GetTmeTermsFromIndex(index int) (tmeTerms []interface{}, error)` - returns a set of terms, having a maximum of `maxRecord` elements starting from the provided index
* `GetTmeTermById(tmeID string) (tmeTerm interface{}, error)` - returns the term details, obtained by the tme term identifier

To run, create a new repository, e.g.:

    NewTimeRepositoryWithConfig(tmeRepositoryConfig{
            client: &client,
            tmeBaseURL: ,
            userName: "username",
            password: "password",
            token: "token",
            maxRecords: 100,
            slices: 1,
            taxonomyName: "GL",
            source: &AuthorityFiles{},
            modelTransformer: new(dummyTransformer),
        })

The modelTransformer should implement the following methods, according to his own model type:

* `UnMarshallTaxonomy(contents []byte) (tmeTerms []interface{}, error)` - loading xml data into list of terms (`[]tmeTerms`)
* `UnMarshallTerm(content []byte) (tmeTerm interface{}, error)` - loading xml data into a tmeTerm model

