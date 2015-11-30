//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"encoding/json"
	"sort"
)

/*
Type Bindings is a slice of pointers to
Binding.
*/
type Bindings []*Binding

/*
Bindings is a helper class. Type Binding is a struct
with three fields, variable, expression and descend.
*/
type Binding struct {
	variable string     `json:"variable"`
	expr     Expression `json:"expr"`
	descend  bool       `json:"descend"`
}

/*
This method returns a pointer to the Binding struct with
input variable and expression used to set the fields of
the structure. The descend boolean field is set to false.
*/
func NewBinding(variable string, expr Expression) *Binding {
	return &Binding{variable, expr, false}
}

/*
This method returns a new binding with the descendant
field for the Binding struct set to true.
*/
func NewDescendantBinding(variable string, expr Expression) *Binding {
	return &Binding{variable, expr, true}
}

func (this *Binding) Copy() *Binding {
	return &Binding{
		variable: this.variable,
		expr:     this.expr.Copy(),
		descend:  this.descend,
	}
}

/*
This method is used to access the variable field
of the receiver which is of type Binding.
*/
func (this *Binding) Variable() string {
	return this.variable
}

/*
This method is used to access the expression field
of the receiver which is of type Binding.
*/
func (this *Binding) Expression() Expression {
	return this.expr
}

/*
This method is used to set the expression field
of the receiver which is of type Binding.
*/
func (this *Binding) SetExpression(expr Expression) {
	this.expr = expr
}

/*
This method is used to access the descend field
of the receiver which is of type Binding.
*/
func (this *Binding) Descend() bool {
	return this.descend
}

func (this *Binding) MarshalJSON() ([]byte, error) {
	r := make(map[string]interface{}, 3)
	r["variable"] = this.variable
	r["expr"] = NewStringer().Visit(this.expr)
	if this.descend {
		r["descend"] = this.descend
	}

	return json.Marshal(r)
}

/*
This method ranges over the bindings (receiver) and maps
each expression to another.
*/
func (this Bindings) MapExpressions(mapper Mapper) (err error) {
	for _, b := range this {
		expr, err := mapper.Map(b.expr)
		if err != nil {
			return err
		}

		b.expr = expr
	}

	return
}

/*
   Returns all contained Expressions.
*/
func (this Bindings) Expressions() Expressions {
	exprs := make(Expressions, len(this))

	for i, b := range this {
		exprs[i] = b.expr
	}

	return exprs
}

func (this Bindings) Copy() Bindings {
	copies := make(Bindings, len(this))
	for i, b := range this {
		copies[i] = b.Copy()
	}

	return copies
}

// Implement sort.Interface

func (this Bindings) Len() int {
	return len(this)
}

func (this Bindings) Less(i, j int) bool {
	return this[i].variable < this[j].variable
}

func (this Bindings) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this Bindings) Sort() {
	sort.Sort(this)
}
