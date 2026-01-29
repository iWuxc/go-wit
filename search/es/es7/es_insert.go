package es7

import (
	"context"
	"github.com/iWuxc/go-wit/search/contract"
	"github.com/pkg/errors"
)

func (e *ESClient) InsertWithRouting(ctx context.Context, index, id, routing string, doc interface{}) error {
	if _, err := e.Client.Index().Index(index).Id(id).Routing(routing).BodyJson(doc).Refresh("true").Do(ctx); err != nil {
		return err
	}

	return nil
}

func (e *ESClient) InsertMultipleWithRoutings(ctx context.Context, index string, ids []string, routings []string, items []interface{}) error {
	m := make([]contract.Multiple, len(ids))
	if len(items) == 0 {
		return errors.New("Docs can't be empty")
	}

	if len(items) != len(ids) {
		return errors.New("Docs count not match ids")
	}
	if len(items) != len(routings) {
		return errors.New("Docs count not match routings")
	}

	for i, id := range ids {
		m = append(m, contract.Multiple{
			Index:   index,
			ID:      id,
			OpType:  contract.OpIndex,
			Doc:     items[i],
			Routing: routings[i],
		})
	}

	return e.Multiple(ctx, m...)
}

func (e *ESClient) Insert(ctx context.Context, index, id string, doc interface{}) error {
	if _, err := e.Client.Index().Index(index).Id(id).BodyJson(doc).Refresh("true").Do(ctx); err != nil {
		return err
	}

	return nil
}

func (e *ESClient) InsertMultiple(ctx context.Context, index string, ids []string, items []interface{}) error {
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
			OpType: contract.OpIndex,
			Doc:    items[i],
		})
	}

	return e.Multiple(ctx, m...)
}
