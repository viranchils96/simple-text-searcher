package main

import (
	"flag"
	"log"
	"time"

	utils "github.com/viranchils96/simple-text-searcher/utils"
)

func main() {
	var path, query string
	flag.StringVar(&path, "p", "enwiki-latest-abstract1.xml.gz", "wiki abstract path")
	flag.StringVar(&query, "q", "Small wild cat", "search query")
	flag.Parse()
	log.Println("-------Search in Progress------")
	start := time.Now()
	docs, err := utils.LoadDocuments(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("load time %d of documents in %v", len(docs), time.Since(start))
	start = time.Now()
	idx := make(utils.Index)
	idx.Add(docs)
	log.Printf("Indexed %d documents in %v", len(docs), time.Since(start))
	start = time.Now()
	matchedIDs := idx.Search(query)
	log.Printf("Search found %d documents in %v", len(matchedIDs), time.Since(start))
	for _, id := range matchedIDs {
		doc := docs[id]
		log.Printf("%d\t%s\n", id, doc.Text)
	}
}
