package contract

type Aggregations interface {
	// Source returns a JSON-serializable aggregation that is a fragment
	// of the request sent to Elasticsearch.
	Source() (interface{}, error)
}
