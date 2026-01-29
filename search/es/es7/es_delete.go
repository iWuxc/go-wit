package es7

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/search/contract"
	"github.com/olivere/elastic/v7"
)

func (e *ESClient) DeleteMultiple(ctx context.Context, index string, ids []string, items []interface{}) error {
	m := make([]contract.Multiple, len(ids))
	if len(items) == 0 {
		return errors.New("Docs can't be empty")
	}

	if len(items) != len(ids) {
		return errors.New("Docs count not match ids")
	}

	for i, id := range ids {
		m = append(m, contract.Multiple{
			Index:  index,
			ID:     id,
			OpType: contract.OpDelete,
			Doc:    items[i],
		})
	}

	return e.Multiple(ctx, m...)
}

func (e *ESClient) DeleteByQuery(ctx context.Context, index string, query contract.ElasticQuery) error {
	_, err := e.Client.DeleteByQuery(index).Query(query).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "DeleteByQuery Do error")
	}

	return nil
}

func (e *ESClient) Delete(ctx context.Context, index string, id string) error {
	_, err := e.Client.Delete().Index(index).Id(id).Refresh("true").Do(ctx)
	if nil != err {
		if elastic.IsNotFound(err) {
			return errors.Wrap(contract.ErrNotFound, "elasticsearch delete by id")
		}
	}

	return errors.Wrap(err, "elasticsearch delete by id")
}
