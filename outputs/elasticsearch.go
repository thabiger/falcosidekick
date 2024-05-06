// SPDX-License-Identifier: Apache-2.0
/*
Copyright (C) 2024 The Falco Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package outputs

import (
	"log"
	"net/url"
	"time"

	"github.com/falcosecurity/falcosidekick/types"
)

// ElasticsearchPost posts event to Elasticsearch
func (c *Client) ElasticsearchPost(falcopayload types.FalcoPayload) {
	c.Stats.Elasticsearch.Add(Total, 1)

	current := time.Now()
	var eURL string
	switch c.Config.Elasticsearch.Suffix {
	case "none":
		eURL = c.Config.Elasticsearch.HostPort + "/" + c.Config.Elasticsearch.Index + "/" + c.Config.Elasticsearch.Type
	case "monthly":
		eURL = c.Config.Elasticsearch.HostPort + "/" + c.Config.Elasticsearch.Index + "-" + current.Format("2006.01") + "/" + c.Config.Elasticsearch.Type
	case "annually":
		eURL = c.Config.Elasticsearch.HostPort + "/" + c.Config.Elasticsearch.Index + "-" + current.Format("2006") + "/" + c.Config.Elasticsearch.Type
	default:
		eURL = c.Config.Elasticsearch.HostPort + "/" + c.Config.Elasticsearch.Index + "-" + current.Format("2006.01.02") + "/" + c.Config.Elasticsearch.Type
	}

	endpointURL, err := url.Parse(eURL)
	if err != nil {
		c.setElasticSearchErrorMetrics()
		log.Printf("[ERROR] : %v - %v\n", c.OutputType, err.Error())
		return
	}

	c.EndpointURL = endpointURL
	if c.Config.Elasticsearch.Username != "" && c.Config.Elasticsearch.Password != "" {
		c.httpClientLock.Lock()
		defer c.httpClientLock.Unlock()
		c.BasicAuth(c.Config.Elasticsearch.Username, c.Config.Elasticsearch.Password)
	}

	for i, j := range c.Config.Elasticsearch.CustomHeaders {
		c.AddHeader(i, j)
	}

	err = c.Post(falcopayload)
	if err != nil {
		c.setElasticSearchErrorMetrics()
		log.Printf("[ERROR] : ElasticSearch - %v\n", err)
		return
	}

	// Setting the success status
	go c.CountMetric(Outputs, 1, []string{"output:elasticsearch", "status:ok"})
	c.Stats.Elasticsearch.Add(OK, 1)
	c.PromStats.Outputs.With(map[string]string{"destination": "elasticsearch", "status": OK}).Inc()
}

// setElasticSearchErrorMetrics set the error stats
func (c *Client) setElasticSearchErrorMetrics() {
	go c.CountMetric(Outputs, 1, []string{"output:elasticsearch", "status:error"})
	c.Stats.Elasticsearch.Add(Error, 1)
	c.PromStats.Outputs.With(map[string]string{"destination": "elasticsearch", "status": Error}).Inc()
}
