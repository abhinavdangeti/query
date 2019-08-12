//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"github.com/couchbase/query/value"
)

/*
Expression that implements array indexing in CREATE INDEX.
*/
type All struct {
	ExpressionBase
	array    Expression
	distinct bool
}

func NewAll(array Expression, distinct bool) *All {
	rv := &All{
		array:    array,
		distinct: distinct,
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *All) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAll(this)
}

func (this *All) Type() value.Type {
	return this.array.Type()
}

func (this *All) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.array.Evaluate(item, context)
}

func (this *All) EvaluateForIndex(item value.Value, context Context) (value.Value, value.Values, error) {
	val, vals, err := this.array.EvaluateForIndex(item, context)
	if err != nil {
		return nil, nil, err
	}

	if vals != nil {
		return val, vals, nil
	}

	var rv value.Values
	switch val.Type() {
	case value.ARRAY:
		act := val.Actual().([]interface{})
		rv = make(value.Values, len(act))
		for i, a := range act {
			rv[i] = value.NewValue(a)
		}
	case value.NULL:
		rv = _NULL_ARRAY
	case value.MISSING:
		rv = _MISSING_ARRAY
	default:
		// Coerce scalar into array
		rv = value.Values{val}
	}

	return val, rv, nil
}

var _NULL_ARRAY = value.Values{value.NULL_VALUE}

var _MISSING_ARRAY = value.Values{value.MISSING_VALUE}

func (this *All) IsArrayIndexKey() (bool, bool) {
	return true, this.distinct
}

func (this *All) Value() value.Value {
	return this.array.Value()
}

func (this *All) Static() Expression {
	return this.array.Static()
}

func (this *All) Alias() string {
	return this.array.Alias()
}

func (this *All) Indexable() bool {
	return this.array.Indexable()
}

func (this *All) PropagatesMissing() bool {
	return this.array.PropagatesMissing()
}

func (this *All) PropagatesNull() bool {
	return this.array.PropagatesNull()
}

func (this *All) EquivalentTo(other Expression) bool {
	all, ok := other.(*All)
	return ok && (this.distinct == all.distinct) &&
		this.array.EquivalentTo(all.array)
}

func (this *All) DependsOn(other Expression) bool {
	// Unwrap other if possible
	for all, ok := other.(*All); ok && (this.distinct || !all.distinct); all, ok = other.(*All) {
		other = all.array
	}

	return this.array.DependsOn(other)
}

func (this *All) CoveredBy(keyspace string, exprs Expressions, options CoveredOptions) Covered {
	return this.array.CoveredBy(keyspace, exprs, options)
}

func (this *All) Children() Expressions {
	return Expressions{this.array}
}

func (this *All) MapChildren(mapper Mapper) error {
	c, err := mapper.Map(this.array)
	if err == nil && c != this.array {
		this.array = c
	}

	return err
}

func (this *All) Copy() Expression {
	rv := NewAll(this.array.Copy(), this.distinct)
	rv.BaseCopy(this)
	return rv
}

func (this *All) Array() Expression {
	return this.array
}

func (this *All) Distinct() bool {
	return this.distinct
}
