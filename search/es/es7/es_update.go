package es7

import (
	"context"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/search/contract"
	elastic7 "github.com/olivere/elastic/v7"
)

func (e *ESClient) UpdateMultiple(ctx context.Context, index string, ids []string, items []interface{}) error {
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

func (e *ESClient) Update(ctx context.Context, index, id string, doc interface{}) error {
	re := "true"

	r, ok := ctx.Value("es_refresh").(bool)
	if ok && !r {
		re = "false"
	}

	_, err := e.Client.Update().Index(index).Id(id).Doc(doc).Refresh(re).Do(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (e *ESClient) UpdateMultipleByOpType(ctx context.Context, index string, opType contract.Op, ids []string, items []interface{}) error {
	m := make([]contract.Multiple, 0, len(ids))
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
			OpType: opType,
			Doc:    items[i],
		})
	}

	return e.Multiple(ctx, m...)
}

func (e *ESClient) UpdateByBulk(ctx context.Context, index string, opType contract.Op, updateRequest []contract.UpdateRequest) error {
	m := make([]contract.Multiple, 0, len(updateRequest))
	if len(updateRequest) == 0 {
		return errors.New("UpdateRequest can't be empty")
	}

	for _, request := range updateRequest {
		if request.Type == contract.UpdateTypeData {
			m = append(m, contract.Multiple{
				Index:      index,
				ID:         request.Id,
				OpType:     opType,
				UpdateType: request.Type,
				Doc:        request.Data,
			})
		} else if request.Type == contract.UpdateTypeScript {
			m = append(m, contract.Multiple{
				Index:      index,
				ID:         request.Id,
				OpType:     opType,
				UpdateType: request.Type,
				Script:     request.ScriptQuery,
			})
		}
	}

	return e.Bulk(ctx, m...)
}

func (e *ESClient) UpdateByScript(ctx context.Context, index, id string, script string) error {
	re := "true"

	r, ok := ctx.Value("es_refresh").(bool)
	if ok && !r {
		re = "false"
	}
	scriptObj := elastic7.NewScript(script).Params(map[string]interface{}{})
	_, err := e.Client.Update().Index(index).Script(scriptObj).Id(id).Refresh(re).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}
