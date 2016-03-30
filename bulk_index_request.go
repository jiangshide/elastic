// Copyright 2012-present Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Bulk request to add a document to Elasticsearch.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
// for details.
type BulkIndexRequest struct {
	BulkableRequest
	index       string
	typ         string
	id          string
	opType      string
	routing     string
	parent      string
	timestamp   string
	ttl         int64
	refresh     *bool
	version     int64  // default is MATCH_ANY
	versionType string // default is "internal"
	doc         interface{}

	source []string
}

// NewBulkIndexRequest returns a new BulkIndexRequest.
// The operation type is "index" by default.
func NewBulkIndexRequest() *BulkIndexRequest {
	return &BulkIndexRequest{
		opType: "index",
	}
}

// Index specifies the Elasticsearch index to use for this index request.
// If unspecified, the index set on the BulkService will be used.
func (r *BulkIndexRequest) Index(index string) *BulkIndexRequest {
	r.index = index
	r.source = nil
	return r
}

// Type specifies the Elasticsearch type to use for this index request.
// If unspecified, the type set on the BulkService will be used.
func (r *BulkIndexRequest) Type(typ string) *BulkIndexRequest {
	r.typ = typ
	r.source = nil
	return r
}

// Id specifies the identifier of the document to index.
func (r *BulkIndexRequest) Id(id string) *BulkIndexRequest {
	r.id = id
	r.source = nil
	return r
}

// OpType specifies if this request should follow create-only or upsert
// behavior. This follows the OpType of the standard document index API.
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html#operation-type
// for details.
func (r *BulkIndexRequest) OpType(opType string) *BulkIndexRequest {
	r.opType = opType
	r.source = nil
	return r
}

// Routing specifies a routing value for the request.
func (r *BulkIndexRequest) Routing(routing string) *BulkIndexRequest {
	r.routing = routing
	r.source = nil
	return r
}

// Parent specifies the identifier of the parent document (if available).
func (r *BulkIndexRequest) Parent(parent string) *BulkIndexRequest {
	r.parent = parent
	r.source = nil
	return r
}

// Timestamp can be used to index a document with a timestamp.
// This is deprecated as of 2.0.0-beta2; you should use a normal date field
// and set its value explicitly.
func (r *BulkIndexRequest) Timestamp(timestamp string) *BulkIndexRequest {
	r.timestamp = timestamp
	r.source = nil
	return r
}

// Ttl (time to live) sets an expiration date for the document. Expired
// documents will be expunged automatically.
// This is deprecated as of 2.0.0-beta2 and will be replaced by a different
// implementation in a future version.
func (r *BulkIndexRequest) Ttl(ttl int64) *BulkIndexRequest {
	r.ttl = ttl
	r.source = nil
	return r
}

// Refresh indicates whether to update the shards immediately after
// the request has been processed. Newly added documents will appear
// in search immediately at the cost of slower bulk performance.
func (r *BulkIndexRequest) Refresh(refresh bool) *BulkIndexRequest {
	r.refresh = &refresh
	r.source = nil
	return r
}

// Version indicates the version of the document as part of an optimistic
// concurrency model.
func (r *BulkIndexRequest) Version(version int64) *BulkIndexRequest {
	r.version = version
	r.source = nil
	return r
}

// VersionType specifies how versions are created. It can be e.g. internal,
// external, external_gte, or force.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html#index-versioning
// for details.
func (r *BulkIndexRequest) VersionType(versionType string) *BulkIndexRequest {
	r.versionType = versionType
	r.source = nil
	return r
}

// Doc specifies the document to index.
func (r *BulkIndexRequest) Doc(doc interface{}) *BulkIndexRequest {
	r.doc = doc
	r.source = nil
	return r
}

// String returns the on-wire representation of the index request,
// concatenated as a single string.
func (r *BulkIndexRequest) String() string {
	lines, err := r.Source()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return strings.Join(lines, "\n")
}

// Source returns the on-wire representation of the index request,
// split into an action-and-meta-data line and an (optional) source line.
// See https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
// for details.
func (r *BulkIndexRequest) Source() ([]string, error) {
	// { "index" : { "_index" : "test", "_type" : "type1", "_id" : "1" } }
	// { "field1" : "value1" }

	if r.source != nil {
		return r.source, nil
	}

	lines := make([]string, 2)

	// We build the JSON via a buffer here to save time in JSON serialization.
	// This is one of the hot paths for bulk indexing.

	// "index" ...
	var comma bool
	var buf bytes.Buffer
	var add = func(k, v string) {
		if comma {
			buf.WriteString(",")
		}
		buf.WriteString(fmt.Sprintf(`%q:%s`, k, v))
		comma = true
	}
	// Keep in alphabetical order to emulate behavior of JSON serializer and tests still pass
	buf.WriteString("{")
	if r.id != "" {
		add("_id", fmt.Sprintf("%q", r.id))
	}
	if r.index != "" {
		add("_index", fmt.Sprintf("%q", r.index))
	}
	if r.parent != "" {
		add("_parent", fmt.Sprintf("%q", r.parent))
	}
	if r.routing != "" {
		add("_routing", fmt.Sprintf("%q", r.routing))
	}
	if r.timestamp != "" {
		add("_timestamp", fmt.Sprintf("%q", r.timestamp))
	}
	if r.ttl > 0 {
		add("_ttl", fmt.Sprintf("%d", r.ttl))
	}
	if r.typ != "" {
		add("_type", fmt.Sprintf("%q", r.typ))
	}
	if r.version > 0 {
		add("_version", fmt.Sprintf("%d", r.version))
	}
	if r.versionType != "" {
		add("_version_type", fmt.Sprintf("%q", r.versionType))
	}
	if r.refresh != nil {
		if *r.refresh {
			add("refresh", "true")
		} else {
			add("refresh", "false")
		}
	}
	buf.WriteString("}")
	lines[0] = fmt.Sprintf(`{"%s":%s}`, r.opType, buf.String())

	// "field1" ...
	if r.doc != nil {
		switch t := r.doc.(type) {
		default:
			body, err := json.Marshal(r.doc)
			if err != nil {
				return nil, err
			}
			lines[1] = string(body)
		case json.RawMessage:
			lines[1] = string(t)
		case *json.RawMessage:
			lines[1] = string(*t)
		case string:
			lines[1] = t
		case *string:
			lines[1] = *t
		}
	} else {
		lines[1] = "{}"
	}

	r.source = lines
	return lines, nil
}
