//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"fmt"
	"github.com/couchbase/cbauth"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/planner"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

type Phases int

const (
	// Execution layer
	AUTHORIZE = Phases(iota)
	FETCH
	INDEX_SCAN
	PRIMARY_SCAN
	JOIN
	INDEX_JOIN
	NL_JOIN
	HASH_JOIN
	NEST
	INDEX_NEST
	NL_NEST
	HASH_NEST
	COUNT
	INDEX_COUNT
	FILTER
	SORT
	INSERT
	DELETE
	UPDATE
	UPSERT
	MERGE
	INFER
	FTS_SEARCH

	// Server layer
	INSTANTIATE
	PARSE
	PLAN
	REPREPARE
	RUN
	PHASES // Sizer
)

func (phase Phases) String() string {
	return _PHASE_NAMES[phase]
}

var _PHASE_NAMES = []string{
	AUTHORIZE:    "authorize",
	FETCH:        "fetch",
	INDEX_SCAN:   "indexScan",
	PRIMARY_SCAN: "primaryScan",
	JOIN:         "join",
	INDEX_JOIN:   "indexJoin",
	NL_JOIN:      "nestedLoopJoin",
	HASH_JOIN:    "hashJoin",
	NEST:         "nest",
	INDEX_NEST:   "indexNest",
	NL_NEST:      "nestedLoopNest",
	HASH_NEST:    "hashNest",
	COUNT:        "count",
	INDEX_COUNT:  "indexCount",
	SORT:         "sort",
	FILTER:       "filter",
	INSERT:       "insert",
	DELETE:       "delete",
	UPDATE:       "update",
	UPSERT:       "upsert",
	MERGE:        "merge",
	INFER:        "inferKeySpace",
	FTS_SEARCH:   "ftsSearch",

	INSTANTIATE: "instantiate",
	PARSE:       "parse",
	PLAN:        "plan",
	REPREPARE:   "reprepare",
	RUN:         "run",
}

const _PHASE_UPDATE_COUNT uint64 = 100

type Output interface {
	SetUp()                                // Any action necessary before processing results
	Result(item value.AnnotatedValue) bool // Process individual items
	CloseResults()                         // Signal that results are through
	Abort(err errors.Error)
	Fatal(err errors.Error)
	Error(err errors.Error)
	Warning(wrn errors.Error)
	AddMutationCount(uint64)
	MutationCount() uint64
	SortCount() uint64
	SetSortCount(i uint64)
	AddPhaseOperator(p Phases)
	AddPhaseCount(p Phases, c uint64)
	FmtPhaseCounts() map[string]interface{}
	FmtPhaseOperators() map[string]interface{}
	AddPhaseTime(phase Phases, duration time.Duration)
	FmtPhaseTimes() map[string]interface{}
}

type Context struct {
	requestId          string
	datastore          datastore.Datastore
	systemstore        datastore.Datastore
	namespace          string
	indexApiVersion    int
	featureControls    uint64
	readonly           bool
	maxParallelism     int
	scanCap            int64
	pipelineCap        int64
	pipelineBatch      int
	isPrepared         bool
	reqDeadline        time.Time
	now                time.Time
	namedArgs          map[string]value.Value
	positionalArgs     value.Values
	credentials        auth.Credentials
	consistency        datastore.ScanConsistency
	scanVectorSource   timestamp.ScanVectorSource
	output             Output
	prepared           *plan.Prepared
	subplans           *subqueryMap
	subresults         *subqueryMap
	httpRequest        *http.Request
	authenticatedUsers auth.AuthenticatedUsers
	mutex              sync.RWMutex
	whitelist          map[string]interface{}
	inlistHashMap      map[*expression.In]*expression.InlistHash
	inlistHashLock     sync.RWMutex
}

func NewContext(requestId string, datastore, systemstore datastore.Datastore,
	namespace string, readonly bool, maxParallelism int, scanCap, pipelineCap int64,
	pipelineBatch int, namedArgs map[string]value.Value, positionalArgs value.Values,
	credentials auth.Credentials, consistency datastore.ScanConsistency,
	scanVectorSource timestamp.ScanVectorSource, output Output, httpRequest *http.Request,
	prepared *plan.Prepared, indexApiVersion int, featureControls uint64) *Context {

	rv := &Context{
		requestId:        requestId,
		datastore:        datastore,
		systemstore:      systemstore,
		namespace:        namespace,
		readonly:         readonly,
		maxParallelism:   maxParallelism,
		scanCap:          scanCap,
		pipelineCap:      pipelineCap,
		pipelineBatch:    pipelineBatch,
		now:              time.Now(),
		namedArgs:        namedArgs,
		positionalArgs:   positionalArgs,
		credentials:      credentials,
		consistency:      consistency,
		scanVectorSource: scanVectorSource,
		output:           output,
		subplans:         nil,
		subresults:       nil,
		httpRequest:      httpRequest,
		prepared:         prepared,
		indexApiVersion:  indexApiVersion,
		featureControls:  featureControls,
		inlistHashMap:    nil,
	}

	if rv.maxParallelism <= 0 || rv.maxParallelism > runtime.NumCPU() {
		rv.maxParallelism = runtime.NumCPU()
	}

	return rv
}

func (this *Context) OriginalHttpRequest() *http.Request {
	return this.httpRequest
}

func (this *Context) RequestId() string {
	return this.requestId
}

func (this *Context) Type() string {
	if this.prepared != nil {
		return this.prepared.Type()
	}
	return ""
}

func (this *Context) Datastore() datastore.Datastore {
	return this.datastore
}

func (this *Context) SetWhitelist(val map[string]interface{}) {
	this.whitelist = val
}

func (this *Context) GetWhitelist() map[string]interface{} {
	return this.whitelist
}

func (this *Context) DatastoreVersion() string {
	return this.datastore.Info().Version()
}

func (this *Context) Systemstore() datastore.Datastore {
	return this.systemstore
}

func (this *Context) Namespace() string {
	return this.namespace
}

func (this *Context) Readonly() bool {
	return this.readonly
}

func (this *Context) MaxParallelism() int {
	return this.maxParallelism
}

func (this *Context) Now() time.Time {
	return this.now
}

func (this *Context) NamedArg(name string) (value.Value, bool) {
	val, ok := this.namedArgs[name]
	return val, ok
}

// The position is 1-based (i.e. 1 is the first position)
func (this *Context) PositionalArg(position int) (value.Value, bool) {
	position--

	if position >= 0 && position < len(this.positionalArgs) {
		return this.positionalArgs[position], true
	} else {
		return nil, false
	}
}

func (this *Context) Credentials() auth.Credentials {
	return this.credentials
}

func (this *Context) UrlCredentials() auth.Credentials {
	// For the cases where the request doesnt have credentials but uses an auth
	// token or x509 certs, we need to derive the credentials to be able to query
	// the fts index.
	dUrl, _ := url.Parse(this.DatastoreURL())
	authenticator := cbauth.Default
	u, p, _ := authenticator.GetHTTPServiceAuth(dUrl.Hostname() + ":" + dUrl.Port())
	return auth.Credentials{u: p}
}

func (this *Context) ScanConsistency() datastore.ScanConsistency {
	return this.consistency
}

func (this *Context) ScanVectorSource() timestamp.ScanVectorSource {
	return this.scanVectorSource
}

// Return []string rather than datastore.AuthenticatedUsers to avoid a circular dependency
// in /expression
func (this *Context) AuthenticatedUsers() []string {
	return this.authenticatedUsers
}

func (this *Context) GetScanCap() int64 {
	if this.scanCap > 0 {
		return this.scanCap
	} else {
		return datastore.GetScanCap()
	}
}

func (this *Context) ScanCap() int64 {
	return this.scanCap
}

func (this *Context) SetScanCap(scanCap int64) {
	this.scanCap = scanCap
}

func (this *Context) GetReqDeadline() time.Time {
	return this.reqDeadline
}

func (this *Context) SetReqDeadline(reqDeadline time.Time) {
	this.reqDeadline = reqDeadline
}

func (this *Context) GetPipelineCap() int64 {
	if this.pipelineCap > 0 {
		return this.pipelineCap
	} else {
		return GetPipelineCap()
	}
}

func (this *Context) PipelineCap() int64 {
	return this.pipelineCap
}

func (this *Context) SetPipelineCap(pipelineCap int64) {
	this.pipelineCap = pipelineCap
}

func (this *Context) GetPipelineBatch() int {
	if this.pipelineBatch > 0 {
		return this.pipelineBatch
	} else {
		return PipelineBatchSize()
	}
}

func (this *Context) PipelineBatch() int {
	return this.pipelineBatch
}

func (this *Context) SetPipelineBatch(pipelineBatch int) {
	this.pipelineBatch = pipelineBatch
}

func (this *Context) IsPrepared() bool {
	return this.isPrepared
}

func (this *Context) SetPrepared(isPrepared bool) {
	this.isPrepared = isPrepared
}

func (this *Context) AddMutationCount(i uint64) {
	this.output.AddMutationCount(i)
}

func (this *Context) MutationCount() uint64 {
	return this.output.MutationCount()
}

func (this *Context) SetSortCount(i uint64) {
	this.output.SetSortCount(i)
}

func (this *Context) SortCount() uint64 {
	return this.output.SortCount()
}

func (this *Context) AddPhaseOperator(p Phases) {
	this.output.AddPhaseOperator(p)
}

func (this *Context) AddPhaseCount(p Phases, c uint64) {
	this.output.AddPhaseCount(p, c)
}

func (this *Context) AddPhaseTime(phase Phases, duration time.Duration) {
	this.output.AddPhaseTime(phase, duration)
}

func (this *Context) SetUp() {
	this.output.SetUp()
}

func (this *Context) Result(item value.AnnotatedValue) bool {
	return this.output.Result(item)
}

func (this *Context) CloseResults() {
	this.output.CloseResults()
}

func (this *Context) Error(err errors.Error) {
	this.output.Error(err)
}

func (this *Context) Abort(err errors.Error) {
	this.output.Abort(err)
}

func (this *Context) Fatal(err errors.Error) {
	this.output.Fatal(err)
}

func (this *Context) Warning(wrn errors.Error) {
	this.output.Warning(wrn)
}

func (this *Context) EvaluateSubquery(query *algebra.Select, parent value.Value) (value.Value, error) {
	subresults := this.getSubresults()
	subresult, ok := subresults.get(query)
	if ok {
		return subresult.(value.Value), nil
	}

	subplans := this.getSubplans()
	subplan, planFound := subplans.get(query)

	if !planFound {
		var err error

		// MB-32140: do not replace named/positional arguments with its value for prepared statements
		namedArgs := this.namedArgs
		positionalArgs := this.positionalArgs
		if this.IsPrepared() {
			namedArgs = nil
			positionalArgs = nil
		}
		subplan, err = planner.Build(query, this.datastore, this.systemstore, this.namespace, true,
			namedArgs, positionalArgs, this.indexApiVersion, this.featureControls)
		if err != nil {
			return nil, err
		}

		// Cache plan
		subplans.set(query, subplan)
	}

	pipeline, err := Build(subplan.(plan.Operator), this)
	if err != nil {
		return nil, err
	}

	// Collect subquery results
	// FIXME: this should handled by the planner
	collect := NewCollect(plan.NewCollect(), this)
	sequence := NewSequence(plan.NewSequence(), this, pipeline, collect)
	sequence.RunOnce(this, parent)

	// Await completion
	collect.waitComplete()

	results := collect.ValuesOnce()
	sequence.Done()

	// Cache results
	if !planFound && !query.IsCorrelated() {
		subresults.set(query, results)
	}

	return results, nil
}

func (this *Context) DatastoreURL() string {
	return this.datastore.URL()
}

func (this *Context) getSubplans() *subqueryMap {
	if this.contextSubplans() == nil {
		this.initSubplans()
	}
	return this.contextSubplans()
}

func (this *Context) initSubplans() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.subplans == nil {
		this.subplans = newSubqueryMap()
	}
}

func (this *Context) contextSubplans() *subqueryMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.subplans
}

func (this *Context) getSubresults() *subqueryMap {
	if this.contextSubresults() == nil {
		this.initSubresults()
	}
	return this.contextSubresults()
}

func (this *Context) contextSubresults() *subqueryMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.subresults
}

func (this *Context) initSubresults() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.subresults == nil {
		this.subresults = newSubqueryMap()
	}
}

// Synchronized map
type subqueryMap struct {
	mutex   sync.RWMutex
	entries map[*algebra.Select]interface{}
}

func newSubqueryMap() *subqueryMap {
	rv := &subqueryMap{}
	rv.entries = make(map[*algebra.Select]interface{})
	return rv
}

func (this *subqueryMap) get(key *algebra.Select) (interface{}, bool) {
	this.mutex.RLock()
	rv, ok := this.entries[key]
	this.mutex.RUnlock()
	return rv, ok
}

func (this *subqueryMap) set(key *algebra.Select, value interface{}) {
	this.mutex.Lock()
	this.entries[key] = value
	this.mutex.Unlock()
}

func (this *Context) assert(test bool, what string) bool {
	if test {
		return true
	}
	logging.Severef("assert failure: %v\n\nrequest text:\n<ud>%v</ud>\n",
		what, this.prepared.Text())
	this.Abort(errors.NewExecutionInternalError(what))
	return false
}

func (this *Context) Recover() {
	err := recover()
	if err != nil {
		buf := make([]byte, 1<<16)
		n := runtime.Stack(buf, false)
		s := string(buf[0:n])
		logging.Severef("panic: %v\n\nrequest text:\n<ud>%v</ud>\n\nstack:\n%v",
			err, this.prepared.Text(), s)

		// TODO - this may very well be a duplicate, if the orchestrator is redirecting
		// the standard error to the same file as the log
		os.Stderr.WriteString(s)
		os.Stderr.Sync()

		this.Abort(errors.NewExecutionPanicError(nil, fmt.Sprintf("Panic: %v", err)))
	}
}

// contextless assert - for when we don't have a context!
// no statement text printend, but behaviour consistent with other asserts
func assert(test bool, what string) bool {
	if test {
		return true
	}
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	s := string(buf[0:n])
	logging.Severef("assert failure: %v\n\nstack:\n%v", what, s)
	return false
}

/*
  The map entry for hash list in the context can be shared among all parallel instances
  of an operator (e.g. Filter), and the same hash table can be shared since all instances
  should have the same expression for the IN-list and we should have already checked that
  the elements of the IN-list are "static".
*/
func (this *Context) GetInlistHash(in *expression.In) *expression.InlistHash {
	this.inlistHashLock.RLock()
	defer this.inlistHashLock.RUnlock()
	if this.inlistHashMap != nil {
		return this.inlistHashMap[in]
	}
	return nil
}

func (this *Context) EnableInlistHash(in *expression.In) {
	if this.inlistHashMap == nil {
		this.inlistHashLock.Lock()
		if this.inlistHashMap == nil {
			this.inlistHashMap = make(map[*expression.In]*expression.InlistHash, 4)
		}
		this.inlistHashLock.Unlock()
	}
	this.inlistHashLock.RLock()
	ih := this.inlistHashMap[in]
	this.inlistHashLock.RUnlock()
	if ih == nil {
		this.inlistHashLock.Lock()
		ih = this.inlistHashMap[in]
		if ih == nil {
			ih = expression.NewInlistHash()
			this.inlistHashMap[in] = ih
		}
		this.inlistHashLock.Unlock()
	}
	ih.EnableHash()
}

func (this *Context) RemoveInlistHash(in *expression.In) {
	this.inlistHashLock.Lock()
	if this.inlistHashMap != nil {
		ih := this.inlistHashMap[in]
		if ih != nil {
			ih.DropHashTab()
			delete(this.inlistHashMap, in)
		}
	}
	this.inlistHashLock.Unlock()
}
