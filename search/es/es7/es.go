package es7

import (
	"context"
	"crypto/tls"
	"github.com/iWuxc/go-wit/errors"
	"github.com/iWuxc/go-wit/log"
	"github.com/iWuxc/go-wit/search/contract"
	"github.com/mitchellh/mapstructure"
	elastic7 "github.com/olivere/elastic/v7"
	"net/http"
	"net/url"
	"time"
)

const (
	TypeElasticSearch7 contract.Type = "elasticsearch7"
)

type ESConfig struct {
	Username  string   `json:"username" mapstructure:"username"`
	Password  string   `json:"password" mapstructure:"password"`
	Endpoints []string `json:"endpoints" mapstructure:"endpoints"`
}

type ESClient struct {
	Client *elastic7.Client
}

// NewElasticsearchEngine creates a new Elasticsearch engine
// with the given configuration. e.g: es://admin:admin@elastic.wxbjq.top:8080
func NewElasticsearchEngine(cfgJSON map[string]interface{}, debug bool) (contract.SearchEngineInterface, error) {
	var cfg ESConfig
	if err := mapstructure.Decode(cfgJSON, &cfg); nil != err {
		log.Errorf("decode elasticsearch configuration error: %s", err.Error())
		return nil, errors.Wrap(err, "decode elasticsearch configuration")
	}

	addHTTPScheme(cfg.Endpoints)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint

	esConfig := []elastic7.ClientOptionFunc{
		elastic7.SetURL(cfg.Endpoints...),
		elastic7.SetSniff(false),
		elastic7.SetErrorLog(log.GetInstance()),
		elastic7.SetInfoLog(log.GetInstance()),
		elastic7.SetHealthcheckInterval(10 * time.Second),
		elastic7.SetBasicAuth(cfg.Username, cfg.Password),
	}

	if debug {
		esConfig = append(esConfig, elastic7.SetTraceLog(log.GetInstance()))
	}

	client, err := elastic7.NewClient(esConfig...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create elasticsearch client")
	}

	// ping connection.
	if len(cfg.Endpoints) == 0 {
		log.Error("please check your configuration with elasticsearch")
		return nil, errors.Wrap(contract.ErrEmptyParam, "elasticsearch broker endpoints empty")
	}

	info, _, err := client.Ping(cfg.Endpoints[0]).Do(context.Background())
	if nil != err {
		log.Errorf("ping elasticsearch %v", err)
		return nil, errors.Wrap(err, "ping elasticsearch")
	}

	log.Debugf("use ElasticsearchDriver version: %+v", info.Version.Number)
	return &ESClient{Client: client}, nil
}

func addHTTPScheme(endpoints []string) []string {
	for index := range endpoints {
		urlIns := url.URL{}
		urlIns.Scheme = "http"
		urlIns.Host = endpoints[index]

		// set endpoint.
		endpoints[index] = urlIns.String()
	}
	return endpoints
}

func (es *ESClient) Bulk(ctx context.Context, multiple ...contract.Multiple) error {
	if len(multiple) == 0 {
		return nil
	}

	refresh := "true"
	if val, ok := ctx.Value("es_refresh").(string); ok {
		refresh = val
	}

	esBuild := es.Client.Bulk().Refresh(refresh)
	for _, m := range multiple {

		if m.ID == "" || m.Index == "" || (m.Doc == nil && m.Script.Script == "") {
			continue
		}

		if len(m.OpType) == 0 {
			m.OpType = contract.OpIndex
		}

		if m.Routing == "" {
			if m.OpType == contract.OpUpdate {
				if m.UpdateType == contract.UpdateTypeData {
					esBuild = esBuild.Add(
						elastic7.NewBulkUpdateRequest().Index(m.Index).Type(m.Index).Id(m.ID).Doc(m.Doc),
					)
				} else if m.UpdateType == contract.UpdateTypeScript {
					script := elastic7.NewScript(m.Script.Script).Params(m.Script.Params)
					esBuild = esBuild.Add(
						elastic7.NewBulkUpdateRequest().Index(m.Index).Id(m.ID).Script(script),
					)
				}
			} else {
				esBuild = esBuild.Add(
					elastic7.NewBulkIndexRequest().OpType(string(m.OpType)).Index(m.Index).Id(m.ID).Doc(m.Doc),
				)
			}
		} else {
			if m.OpType == contract.OpUpdate {
				if m.UpdateType == contract.UpdateTypeData {
					esBuild = esBuild.Add(
						elastic7.NewBulkUpdateRequest().Index(m.Index).Id(m.ID).Routing(m.Routing).Doc(m.Doc),
					)
				} else if m.UpdateType == contract.UpdateTypeScript {
					script := elastic7.NewScript(m.Script.Script).Params(m.Script.Params)
					esBuild = esBuild.Add(
						elastic7.NewBulkUpdateRequest().Index(m.Index).Id(m.ID).Routing(m.Routing).Script(script),
					)
				}
			} else {
				esBuild = esBuild.Add(
					elastic7.NewBulkIndexRequest().OpType(string(m.OpType)).Index(m.Index).Id(m.ID).Routing(m.Routing).Doc(m.Doc),
				)
			}
		}
	}

	_, err := esBuild.Do(ctx)
	if err != nil {
		return err
	}

	return nil
}
