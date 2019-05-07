//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.
package testfs

import (
	"github.com/couchbase/query/errors"
	js "github.com/couchbase/query/test/filestore"
)

func start() *js.MockServer {
	return js.Start("dir:", "../../../data/", js.Namespace_FS)
}

func testCaseFile(fname string, qc *js.MockServer) (fin_stmt string, errstring error) {
	fin_stmt, errstring = js.FtestCaseFile(fname, qc, js.Namespace_FS)
	return
}

func Run_test(mockServer *js.MockServer, q string) ([]interface{}, []errors.Error, errors.Error) {
	return js.Run(mockServer, true, q, nil, nil, js.Namespace_FS)
}
