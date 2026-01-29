package es7

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/search/contract"
	elastic7 "github.com/olivere/elastic/v7"
)

// Search .
func (es *ESClient) Count(ctx context.Context, index string, req contract.SearchRequest) (int64, error) {

	boolQuery := elastic7.NewBoolQuery()
	boolQuery = boolQuery.MinimumNumberShouldMatch(req.MinimumNumberShouldMatch)
	countQuery := es.Client.Count().Index(index)

	for text, fields := range req.Query {
		if len(text) == 0 {
			continue
		}
		boolQuery = boolQuery.Must(elastic7.NewMultiMatchQuery(text, fields...).Type("best_fields"))
	}

	if req.Condition != nil {
		boolQuery = condition2boolQuery(req.Condition, boolQuery)
	}

	countQuery = countQuery.Query(boolQuery)

	countResult, err := countQuery.Pretty(true).Do(ctx)
	if err != nil {
		return countResult, errors.Wrap(err, "query count failed")
	}

	return countResult, nil
}
