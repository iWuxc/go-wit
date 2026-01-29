package es7

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/search/contract"
	"github.com/iWuxc/go-wit/utils"
	elastic7 "github.com/olivere/elastic/v7"
	"reflect"
)

func (e *ESClient) Search(ctx context.Context, index string, req contract.SearchRequest) (contract.SearchResponse, error) {
	resp := contract.SearchResponse{}
	boolQuery := elastic7.NewBoolQuery()
	boolQuery = boolQuery.MinimumNumberShouldMatch(req.MinimumNumberShouldMatch)
	searchQuery := e.Client.Search().Index(index)

	for text, fields := range req.Query {
		if len(text) == 0 {
			continue
		}
		boolQuery = boolQuery.Must(elastic7.NewMultiMatchQuery(text, fields...).Type("best_fields"))
	}

	if req.Condition != nil {
		boolQuery = condition2boolQuery(req.Condition, boolQuery)
	}

	searchQuery = searchQuery.Query(boolQuery)
	if req.Pager = defaultPage(req.Pager); req.Pager.PageSize > 0 {
		searchQuery = searchQuery.From(req.Pager.PageSize * (req.Pager.Page - 1)).Size(req.Pager.PageSize)
	}

	for _, by := range req.OrderBy {
		searchQuery = searchQuery.SortWithInfo(elastic7.SortInfo{
			Field:          by.Field,
			Ascending:      by.Order == contract.OrderByAsc,
			Missing:        by.Missing,
			IgnoreUnmapped: by.IgnoreUnmapped,
			UnmappedType:   by.UnmappedType,
			SortMode:       by.SortMode,
			NestedFilter:   by.NestedFilter,
			Filter:         by.Filter,
			NestedPath:     by.NestedPath,
			Path:           by.Path,
		})
	}

	searchQuery.FetchSourceContext(elastic7.NewFetchSourceContext(req.FetchRaw).Include(req.Fields...))

	if len(req.Aggs) > 0 {
		//for key, ag := range req.Aggs {
		//	if ag != nil {
		//		searchQuery.Aggregation(key, ag)
		//	}
		//}
		return resp, errors.New("es7暂不支持聚合查询")
	}

	searchResult, err := searchQuery.Pretty(true).Do(ctx)
	if err != nil {
		return resp, errors.Wrap(err, "query search failed")
	}

	var data []map[string]interface{}
	for _, item := range searchResult.Each(reflect.TypeOf(map[string]interface{}{})) {
		if t, ok := item.(map[string]interface{}); ok {
			// Filter out any attributes that are not fields in the struct
			if req.Filter != nil {
				req.Filter(t)
			}

			data = append(data, t)
		}
	}

	if searchResult.Hits != nil && searchResult.Hits.Hits != nil && len(searchResult.Hits.Hits) != 0 {
		for _, hit := range searchResult.Hits.Hits {
			resp.IDS = append(resp.IDS, hit.Id)
		}
	}

	resp.Total = searchResult.TotalHits()
	resp.Data = data
	resp.Raw, _ = utils.Marshal(data)
	//resp.Aggregations = searchResult.Aggregations // es7暂不支持聚合，可以使用原生方法替代查询
	if req.Pager != nil {
		resp.Page = req.Pager.Page
		resp.PageSize = req.Pager.PageSize
	}

	return resp, nil
}

// reference: https://www.tutorialspoint.com/elasticsearch/elasticsearch_query_dsl.htm#:~:text=In%20Elasticsearch%2C%20searching%20is%20carried%20out%20by%20using,look%20for%20a%20specific%20value%20in%20specific%20field.
// convert condition.
func condition2boolQuery(conditions []*contract.SearchCondition, boolQuery *elastic7.BoolQuery) *elastic7.BoolQuery {
	for _, condition := range conditions {
		var subQuery *elastic7.BoolQuery
		if condition.SubCondition != nil {
			subQuery = condition2boolQuery(condition.SubCondition, elastic7.NewBoolQuery())
		}

		query := operation(condition)
		switch condition.Clause {
		default:
			if subQuery != nil {
				boolQuery = boolQuery.Must(query, subQuery)
			} else {
				boolQuery = boolQuery.Must(query)
			}

		case contract.ClauseMustNot:
			if subQuery != nil {
				boolQuery = boolQuery.MustNot(query, subQuery)
			} else {
				boolQuery = boolQuery.MustNot(query)
			}
		case contract.ClauseShould:
			if subQuery != nil {
				boolQuery = boolQuery.Should(query, subQuery)
			} else {
				boolQuery = boolQuery.Should(query)
			}
		case contract.ClauseFilter:
			if subQuery != nil {
				boolQuery = boolQuery.Filter(query, subQuery)
			} else {
				boolQuery = boolQuery.Filter(query)
			}
		}
	}
	return boolQuery
}

// defaultPage .
func defaultPage(page *contract.Pager) *contract.Pager {
	if nil == page {
		page = &contract.Pager{}
	}

	if page.Page <= 0 {
		page.Page = 1
	}

	if page.PageSize == 0 {
		page.PageSize = page.DefaultPageSize
	}

	return page
}

// operation convert condition.
func operation(condition *contract.SearchCondition) elastic7.Query {
	if condition.NestedQuery != nil {
		return condition.NestedQuery
	}
	if condition.HasChildQuery != nil {
		return condition.HasChildQuery
	}
	if condition.BoolQuery != nil {
		return condition.BoolQuery
	}

	var q elastic7.Query

	switch condition.Operator {
	case contract.SearchOpLt:
		q = elastic7.NewRangeQuery(condition.Field).Lt(condition.Value)

	case contract.SearchOpLte:
		q = elastic7.NewRangeQuery(condition.Field).Lte(condition.Value)

	case contract.SearchOpGt:
		q = elastic7.NewRangeQuery(condition.Field).Gt(condition.Value)

	case contract.SearchOpGte:
		q = elastic7.NewRangeQuery(condition.Field).Gte(condition.Value)

	case contract.SearchOpNeq:
		q = elastic7.NewTermQuery(condition.Field+".keyword", condition.Value)

	case contract.SearchOpEq, contract.SearchOpTerm:
		switch condition.Value.(type) {
		case []interface{}:
			q = elastic7.NewTermsQuery(condition.Field, condition.Value.([]interface{})...)
		case string:
			q = elastic7.NewTermQuery(condition.Field, condition.Value.(string))
		default:
			q = elastic7.NewTermsQuery(condition.Field, condition.Value)
		}

	case contract.SearchOpPrefix:
		if _, ok := condition.Value.(string); ok {
			q = elastic7.NewPrefixQuery(condition.Field, condition.Value.(string))
		}

	case contract.SearchOpWildcard:
		if _, ok := condition.Value.(string); ok {
			q = elastic7.NewWildcardQuery(condition.Field, "*"+condition.Value.(string)+"*")
		}

	case contract.SearchOpIn:
		if _, ok := condition.Value.([]interface{}); ok {
			q = elastic7.NewTermsQuery(condition.Field, condition.Value.([]interface{})...)
		}

	case contract.SearchOpRegexp:
		if _, ok := condition.Value.(string); ok {
			q = elastic7.NewRegexpQuery(condition.Field+".keyword", condition.Value.(string))
		}

	case contract.SearchOpExists:
		q = elastic7.NewExistsQuery(condition.Field)

	case contract.SearchOpFuzzy:
		if _, ok := condition.Value.(string); ok {
			q = elastic7.NewFuzzyQuery(condition.Field, condition.Value.(string))
		}

	case contract.SearchOpType:
		if _, ok := condition.Value.(string); ok {
			q = elastic7.NewTypeQuery(condition.Value.(string))
		}

	case contract.SearchOpIDS:
		switch condition.Value.(type) {
		case []string:
			q = elastic7.NewIdsQuery(condition.Value.([]string)[0]).Ids(condition.Value.([]string)[1:]...)
		case string:
			q = elastic7.NewIdsQuery().Ids(condition.Value.(string))
		}

	case contract.SearchOpMatch:
		switch condition.Value.(type) {
		case []string:
			q = elastic7.NewMultiMatchQuery(condition.Value.([]string)[0], condition.Value.([]string)[1:]...)
		case string:
			q = elastic7.NewMultiMatchQuery(condition.Value.(string), condition.Field)
		}
	}
	if q == nil {
		q = elastic7.NewMatchQuery(condition.Field, condition.Value)
	}

	if condition.ConstantScore {
		q = elastic7.NewConstantScoreQuery(q).Boost(condition.Boost)
	}

	return q
}
