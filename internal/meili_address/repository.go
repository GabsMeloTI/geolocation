package meiliaddress

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/meilisearch/meilisearch-go"
)

type InterfaceRepository interface {
	FindAddresses(ctx context.Context, query string, limit int) ([]MeiliAddress, error)
}

type Repository struct {
	meili meilisearch.IndexManager
}

func NewMeiliAddressRepository(meiliURL, apiKey string) *Repository {
	client := meilisearch.New(meiliURL, meilisearch.WithAPIKey(apiKey))

	return &Repository{
		meili: client.Index("addresses"),
	}
}

func (r *Repository) FindAddresses(ctx context.Context, query string, limit int) ([]MeiliAddress, error) {
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

	var results []MeiliAddress
	if err := json.Unmarshal(hitsJSON, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search hits: %w", err)
	}

	fmt.Printf("🔎 Meilisearch: %d/%d results returned (limit: %d)\n",
		len(results), searchResp.EstimatedTotalHits, limit)

	return results, nil
}
