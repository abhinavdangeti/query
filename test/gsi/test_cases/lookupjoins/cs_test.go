//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.
package lookupjoins

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLookupjoins(t *testing.T) {
	if strings.ToLower(os.Getenv("GSI_TEST")) != "true" {
		return
	}

	qc := start_cs()

	fmt.Println("\n\nInserting values into Bucket \n\n ")
	runMatch("insert.json", false, false, qc, t)

	runStmt(qc, "CREATE PRIMARY INDEX ON purchase")
	runStmt(qc, "CREATE PRIMARY INDEX ON customer")
	runStmt(qc, "CREATE PRIMARY INDEX ON product")

	runMatch("case_innerjoin.json", false, false, qc, t)
	runMatch("case_joins.json", false, false, qc, t)
	runMatch("case_leftjoin.json", false, false, qc, t)

	_, _, errcs := runStmt(qc, "delete from purchase where test_id = \"joins\"")
	if errcs != nil {
		t.Errorf("did not expect err %s", errcs.Error())
	}

	_, _, errcs = runStmt(qc, "delete from customer where test_id = \"joins\"")
	if errcs != nil {
		t.Errorf("did not expect err %s", errcs.Error())
	}

	_, _, errcs = runStmt(qc, "delete from product where test_id = \"joins\"")
	if errcs != nil {
		t.Errorf("did not expect err %s", errcs.Error())
	}

	runStmt(qc, "DROP PRIMARY INDEX ON purchase")
	runStmt(qc, "DROP PRIMARY INDEX ON customer")
	runStmt(qc, "DROP PRIMARY INDEX ON product")
}