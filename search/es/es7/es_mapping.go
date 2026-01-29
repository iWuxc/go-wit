package es7

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/search/contract"
	"github.com/olivere/elastic/v7"
)

type MappingModel struct {
	Mappings Mapping `json:"mappings"`
	Settings `json:"settings"`
}

type Mapping struct {
	Dynamic    bool                              `json:"dynamic"` // false
	Properties map[string]map[string]interface{} `json:"properties"`
}

type Settings struct {
	MaxResultWindow             int   `json:"max_result_window"`              // 1000
	NumberOfReplicas            int64 `json:"number_of_replicas"`             // 1
	NumberOfShards              int64 `json:"number_of_shards"`               // 1
	IndexMappingIgnoreMalformed bool  `json:"index.mapping.ignore_malformed"` // true
}

func (e *ESClient) Multiple(ctx context.Context, multiple ...contract.Multiple) error {
	if len(multiple) == 0 {
		return nil
	}

	refresh := "true"
	if val, ok := ctx.Value("es_refresh").(string); ok {
		refresh = val
	}

	esBuild := e.Client.Bulk().Refresh(refresh)
	for _, m := range multiple {

		if m.ID == "" || m.Doc == nil || m.Index == "" {
			continue
		}

		if len(m.OpType) == 0 {
			m.OpType = contract.OpIndex
		}

		if m.Routing == "" {
			esBuild = esBuild.Add(
				elastic.NewBulkIndexRequest().OpType(string(m.OpType)).Index(m.Index).Id(m.ID).Doc(m.Doc),
			)
		} else {
			esBuild = esBuild.Add(
				elastic.NewBulkIndexRequest().OpType(string(m.OpType)).Index(m.Index).Id(m.ID).Routing(m.Routing).Doc(m.Doc),
			)
		}

	}

	_, err := esBuild.Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (e *ESClient) BuildIndex(ctx context.Context, index, content string) error {
	if exist, _ := e.Client.IndexExists(index).Do(ctx); exist {
		return nil
	}

	if _, err := e.Client.CreateIndex(index).BodyString(content).Do(ctx); err != nil {
		return errors.Wrap(err, "creating index in es error")
	}

	return nil
}

func (e *ESClient) DeleteIndex(ctx context.Context, index string) error {
	if exist, _ := e.Client.IndexExists(index).Do(ctx); !exist {
		return nil
	}

	if _, err := e.Client.DeleteIndex(index).Do(ctx); err != nil {
		return errors.Wrap(err, "deleting index in es error")
	}

	return nil
}
