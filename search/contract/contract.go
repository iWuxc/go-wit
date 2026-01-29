package contract

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/olivere/elastic"
)

var (
	ErrEmptyParam = errors.New("elasticsearch empty params")
	ErrNotFound   = errors.New("elasticsearch not found")
)

type ElasticQuery interface {
	// Source returns the JSON-serializable query request.
	Source() (interface{}, error)
}

// SearchEngineInterface is the interface that wraps the basic Search method.
type SearchEngineInterface interface {

	// BuildIndex builds the index for the given content string.
	BuildIndex(ctx context.Context, index, content string) error

	// DeleteIndex deletes the index for the given index name.
	DeleteIndex(ctx context.Context, index string) error

	// Search searches the index for the given query string.
	Search(ctx context.Context, index string, request SearchRequest) (SearchResponse, error)

	// Multiple operation Multiple data on ES.
	Multiple(ctx context.Context, multiple ...Multiple) error

	// Insert inserts the given document into the index.
	Insert(ctx context.Context, index, id string, doc interface{}) error

	InsertWithRouting(ctx context.Context, index, id, routing string, doc interface{}) error

	// InsertMultiple inserts the given documents into the index.
	InsertMultiple(ctx context.Context, index string, ids []string, items []interface{}) error

	UpdateMultipleByOpType(ctx context.Context, index string, opType Op, ids []string, items []interface{}) error

	UpdateByBulk(ctx context.Context, index string, opType Op, updateRequest []UpdateRequest) error

	InsertMultipleWithRoutings(ctx context.Context, index string, ids []string, routings []string, items []interface{}) error

	// Update updates the given document in the index.
	Update(ctx context.Context, index, id string, doc interface{}) error

	// UpdateByScript updates a document in Elasticsearch, supported script.
	UpdateByScript(ctx context.Context, index, id string, script string) error

	// UpdateMultiple updates the given documents into the index.
	UpdateMultiple(ctx context.Context, index string, ids []string, items []interface{}) error

	// Delete delete document in elasticsearch
	Delete(ctx context.Context, index string, id string) error

	// DeleteMultiple delete the given documents into the index.
	DeleteMultiple(ctx context.Context, index string, ids []string, items []interface{}) error

	DeleteByQuery(ctx context.Context, index string, query ElasticQuery) error

	Count(ctx context.Context, index string, request SearchRequest) (int64, error)
}

type Op string

const (
	OpIndex  Op = "index"
	OpCreate Op = "create"
	OpUpdate Op = "update"
	OpDelete Op = "delete"
)

type Multiple struct {
	Index      string      `json:"index"`
	ID         string      `json:"id"`
	OpType     Op          `json:"op_type"`
	Doc        interface{} `json:"doc"`
	Routing    string      `json:"routing"`
	UpdateType UpdateType  `json:"update_type"`
	Script     ScriptQuery `json:"script"`
}

type (
	Filter   func(map[string]interface{})
	Type     string
	Operator string
	Order    string
	Clause   string
)

// SearchRequest is the request for search
type SearchRequest struct {
	Filter                   Filter                         `json:"filter"`              // 结果过滤
	FetchRaw                 bool                           `json:"fetch_raw"`           // 是否返回原始数据
	Fields                   []string                       `json:"fields"`              // 返回的字段
	Query                    map[string][]string            `json:"query,omitempty"`     // 查询关键词 map[关键词][]string{字段1, 字段2}
	Pager                    *Pager                         `json:"page,omitempty"`      // 分页信息
	Condition                []*SearchCondition             `json:"condition,omitempty"` // 查询条件
	OrderBy                  []OrderBy                      `json:"order_by,omitempty"`  // 排序
	MinimumNumberShouldMatch int                            `json:"minimum_number_should_match,omitempty"`
	Aggs                     map[string]elastic.Aggregation `json:"aggs"`
}

// SearchResponse is the response for search
type SearchResponse struct {
	Total        int64                    `json:"total,omitempty"`
	IDS          []string                 `json:"ids,omitempty"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Raw          []byte                   `json:"raw,omitempty"`
	Aggregations elastic.Aggregations     `json:"aggregations,omitempty"`
	PageSize     int                      `json:"page_size,omitempty"`
	Page         int                      `json:"page,omitempty"`
}

// CountResponse is the response for count
type CountResponse struct {
	Count int64 `json:"count"`
}

// Pager is used to paginate the results of a search.
type Pager struct {
	Page            int `json:"page"`
	PageSize        int `json:"page_size"`
	DefaultPageSize int `json:"default_page_size"`
}

// OrderBy .
type OrderBy struct {
	Field          string `json:"field"`
	Order          Order  `json:"order"`
	Missing        interface{}
	IgnoreUnmapped *bool
	UnmappedType   string
	SortMode       string
	NestedFilter   Query // deprecated in 6.1 and replaced by Filter
	Filter         Query
	NestedPath     string // deprecated in 6.1 and replaced by Path
	Path           string
}

type Query interface {
	// Source returns the JSON-serializable query request.
	Source() (interface{}, error)
}

const (
	OrderByAsc  Order = "asc"
	OrderByDesc Order = "desc"
)

const (
	ClauseMust    Clause = "must"
	ClauseMustNot Clause = "must_not"
	ClauseShould  Clause = "should"
	ClauseFilter  Clause = "filter"
)

const (
	SearchOpLt       Operator = "$lt"
	SearchOpLte      Operator = "$lte"
	SearchOpGt       Operator = "$gt"
	SearchOpGte      Operator = "$gte"
	SearchOpNeq      Operator = "$neq"
	SearchOpEq       Operator = "$eq"
	SearchOpPrefix   Operator = "$prefix"
	SearchOpWildcard Operator = "$wildcard"
	SearchOpRegexp   Operator = "$regexp"
	SearchOpFuzzy    Operator = "$fuzzy"
	SearchOpType     Operator = "$type"
	SearchOpIDS      Operator = "$ids"
	SearchOpIn       Operator = "$in"
	SearchOpExists   Operator = "$exists"
	SearchOpMatch    Operator = "$match"
	SearchOpTerm     Operator = "$term"
	SearchOpBool     Operator = "$bool"
)

// SearchCondition	is used to filter the results of a search.
type SearchCondition struct {
	Field            string                    `json:"field"`
	Value            interface{}               `json:"value"`
	Clause           Clause                    `json:"clause"`
	Operator         Operator                  `json:"operator"`
	SubCondition     []*SearchCondition        `json:"sub_condition"`
	ScriptQuery      *elastic.ScriptQuery      `json:"script_query"`
	NestedQuery      *elastic.NestedQuery      `json:"nested_query"`
	HasChildQuery    *elastic.HasChildQuery    `json:"has_child_query"`
	BoolQuery        *elastic.BoolQuery        `json:"bool_query"`
	ContextSuggester *elastic.ContextSuggester `json:"context_suggester"`
	ConstantScore    bool                      `json:"constant_score"`
	Boost            float64                   `json:"boost"` // constant_score为true时，boost字段的值才生效
}

// 设置积分排序
type FieldValueWeight struct {
	Field  string  `json:"field"`
	Value  int     `json:"value"`
	Weight float64 `json:"weight"`
}

type ScriptQuery struct {
	Script string                 `json:"script"`
	Params map[string]interface{} `json:"params"`
}
type UpdateType string

const (
	UpdateTypeScript UpdateType = "script"
	UpdateTypeData   UpdateType = "data"
)

type UpdateRequest struct {
	Id          string      `json:"id"`
	Type        UpdateType  `json:"type"`
	ScriptQuery ScriptQuery `json:"script_query"`
	Data        interface{} `json:"data"`
}
