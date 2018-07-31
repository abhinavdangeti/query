//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/value"
)

/*
Represents the Execute command. The argument to EXECUTE must
evaluate to a prepared statement or a string. Type Execute
is a struct that contains a json object value that represents
a plan.Prepared.
*/
type Execute struct {
	statementBase

	prepared value.Value `json:"prepared"`

	// this contains either named parameters (a map of values)
	// or positional (a slice)
	using expression.Expression `json:"using"`
}

/*
The function NewExecute returns a pointer to the Execute
struct with the input argument expressions value as a field.
*/
func NewExecute(prepared expression.Expression, using expression.Expression) *Execute {
	var preparedValue value.Value

	switch prepared := prepared.(type) {
	case *expression.Identifier:
		preparedValue = value.NewValue(prepared.Alias())
	default:
		preparedValue = prepared.Value()
	}

	rv := &Execute{
		prepared: preparedValue,
		using:    using,
	}

	rv.stmt = rv
	return rv
}

/*
It calls the VisitExecute method by passing in the receiver
and returns the interface. It is a visitor pattern.
*/
func (this *Execute) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitExecute(this)
}

/*
This method returns the shape of the result, which is
the signature of the input prepared statement.
*/
func (this *Execute) Signature() value.Value {
	signature, _ :=
		this.prepared.Field("signature")
	return signature
}

/*
It's an execute
*/
func (this *Execute) Type() string {
	return "EXECUTE"
}

/*
Returns nil.
*/
func (this *Execute) Formalize() error {
	return nil
}

/*
Returns nil.
*/
func (this *Execute) MapExpressions(mapper expression.Mapper) error {
	return nil
}

/*
Returns all contained Expressions.
*/
func (this *Execute) Expressions() expression.Expressions {
	return nil
}

/*
Returns all required privileges.
*/
func (this *Execute) Privileges() (*auth.Privileges, errors.Error) {
	return nil, nil
}

/*
Returns the input prepared value that represents the prepared
statement.
*/
func (this *Execute) Prepared() value.Value {
	return this.prepared
}

/*
Returns the input placeholder values
*/
func (this *Execute) Using() expression.Expression {
	return this.using
}
