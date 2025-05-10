package meiliaddress

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
)

type InterfaceRepository interface {
	FindMeiliStreetsRepository(ctx context.Context, query string, limit int) ([]MeiliStreets, error)
}

type Repository struct {
	meili meilisearch.IndexManager
}

func NewMeiliAddressRepository(meiliURL, apiKey string) *Repository {
	client := meilisearch.New(meiliURL, meilisearch.WithAPIKey(apiKey))

	return &Repository{
		meili: client.Index("streets"),
	}
}

func (r *Repository) FindMeiliStreetsRepository(query string, limit int) ([]MeiliStreets, error) {
	searchResp, err := r.meili.Search(query, &meilisearch.SearchRequest{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("error querying meilisearch: %w", err)
	}

	hitsJSON, err := json.Marshal(searchResp.Hits)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search hits: %w", err)
	}

	var results []MeiliStreets
	if err := json.Unmarshal(hitsJSON, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search hits: %w", err)
	}

	return results, nil
}
