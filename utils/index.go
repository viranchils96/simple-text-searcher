package utils

import (
	"math"
	"sort"
	"sync"
)

type IndexShard struct {
	sync.RWMutex
	data map[string][]int
	tf   map[string]map[int]float32
}

type Index struct {
	shards []*IndexShard
	count  int
}

type SearchResult struct {
	ID    int
	Score float32
	Text  string
}

func NewIndex(shardCount int) *Index {
	shards := make([]*IndexShard, shardCount)
	for i := range shards {
		shards[i] = &IndexShard{
			data: make(map[string][]int),
			tf:   make(map[string]map[int]float32),
		}
	}
	return &Index{shards: shards, count: shardCount}
}

func (idx *Index) getShard(term string) *IndexShard {
	h := fnv32(term) % uint32(idx.count)
	return idx.shards[h]
}

func fnv32(s string) uint32 {
	h := uint32(2166136261)
	for _, c := range s {
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}

func (idx *Index) Add(docs []Document) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for _, doc := range docs {
		wg.Add(1)
		sem <- struct{}{}

		go func(d Document) {
			defer wg.Done()
			defer func() { <-sem }()

			tokens := analyze(d.Text)
			tf := make(map[string]float32)

			for _, t := range tokens {
				tf[t] += 1.0 / float32(len(tokens))
			}

			for term, freq := range tf {
				shard := idx.getShard(term)
				shard.Lock()
				shard.data[term] = append(shard.data[term], d.ID)
				if shard.tf[term] == nil {
					shard.tf[term] = make(map[int]float32)
				}
				shard.tf[term][d.ID] = freq
				shard.Unlock()
			}
		}(doc)
	}
	wg.Wait()
}

func (idx *Index) Search(query string, maxResults int, docs []Document) []SearchResult {
	terms := analyze(query)
	results := make(chan SearchResult, 100)
	var wg sync.WaitGroup

	for _, term := range terms {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			shard := idx.getShard(t)
			shard.RLock()
			defer shard.RUnlock()

			if ids, exists := shard.data[t]; exists {
				idf := math.Log(float64(len(docs)) / float64(len(ids)+1))
				for _, id := range ids {
					results <- SearchResult{
						ID:    id,
						Score: float32(idf) * shard.tf[t][id],
					}
				}
			}
		}(term)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	scores := make(map[int]float32)
	for res := range results {
		scores[res.ID] += res.Score
	}

	ranked := make([]SearchResult, 0, len(scores))
	for id, score := range scores {
		ranked = append(ranked, SearchResult{
			ID:    id,
			Score: score,
			Text:  docs[id].Text,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	if len(ranked) > maxResults {
		return ranked[:maxResults]
	}
	return ranked
}
