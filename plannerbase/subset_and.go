//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plannerbase

import (
	"github.com/couchbase/query/expression"
)

func (this *subset) VisitAnd(expr *expression.And) (interface{}, error) {
	expr2 := this.expr2
	value2 := expr2.Value()
	if value2 != nil {
		return value2.Truth(), nil
	}

	if expr.EquivalentTo(expr2) {
		return true, nil
	}

	for _, child := range expr.Operands() {
		if SubsetOf(child, expr2) {
			return true, nil
		}
	}

	switch expr2 := expr2.(type) {
	case *expression.And:
		for _, child2 := range expr2.Operands() {
			if !SubsetOf(expr, child2) {
				return false, nil
			}
		}

		return true, nil
	case *expression.Or:
		for _, child2 := range expr2.Operands() {
			if SubsetOf(expr, child2) {
				return true, nil
			}
		}
	}

	return false, nil
}
