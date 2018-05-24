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
	"encoding/json"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
)

/*
Represents the keyspace-ref used in DML statements. It
contains three fields namespace, keyspace (bucket) and
an alias (as).
*/
type KeyspaceRef struct {
	namespace string `json:"namespace"`
	keyspace  string `json:"keyspace"`
	as        string `json:"as"`
}

/*
The function NewKeyspaceRef returns a pointer to the
KeyspaceRef struct by assigning the input attributes
to the fields of the struct.
*/
func NewKeyspaceRef(namespace, keyspace, as string) *KeyspaceRef {
	return &KeyspaceRef{namespace, keyspace, as}
}

/*
Qualify identifiers for the keyspace. It also makes sure that the
keyspace term contains a name or alias.
*/
func (this *KeyspaceRef) Formalize() (f *expression.Formalizer, err error) {
	keyspace := this.Alias()
	if keyspace == "" {
		err = errors.NewNoTermNameError("Keyspace", "semantics.keyspace.reference_requires_name_or_alias")
		return
	}

	f = expression.NewFormalizer(keyspace, nil)
	return
}

/*
Returns the namespace string.
*/
func (this *KeyspaceRef) Namespace() string {
	return this.namespace

}

/*
Set the default namespace.
*/
func (this *KeyspaceRef) SetDefaultNamespace(namespace string) {
	if this.namespace == "" {
		this.namespace = namespace
	}
}

/*
Returns the keyspace string.
*/
func (this *KeyspaceRef) Keyspace() string {
	return this.keyspace
}

/*
Returns the AS alias string.
*/
func (this *KeyspaceRef) As() string {
	return this.as
}

/*
Returns the alias as the keyspace or the as string
based on if as is empty.
*/
func (this *KeyspaceRef) Alias() string {
	if this.as != "" {
		return this.as
	} else {
		return this.keyspace
	}
}

/*
Marshals input into byte array.
*/
func (this *KeyspaceRef) MarshalJSON() ([]byte, error) {
	r := make(map[string]interface{}, 3)
	r["keyspace"] = this.keyspace
	r["namespace"] = this.namespace
	if this.as != "" {
		r["as"] = this.as
	}

	return json.Marshal(r)
}

/*
Returns the full keyspace name, including the namespace.
*/
func (this *KeyspaceRef) FullName() string {
	return this.namespace + ":" + this.keyspace
}
