//  Copyright (c) 2018 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package server

import (
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/logging"
)

const (
	CPUPROFILE      = "cpuprofile"
	DEBUG           = "debug"
	KEEPALIVELENGTH = "keep-alive-length"
	LOGLEVEL        = "loglevel"
	MAXPARALLELISM  = "max-parallelism"
	MEMPROFILE      = "memprofile"
	REQUESTSIZECAP  = "request-size-cap"
	PIPELINEBATCH   = "pipeline-batch"
	PIPELINECAP     = "pipeline-cap"
	SCANCAP         = "scan-cap"
	SERVICERS       = "servicers"
	TIMEOUTSETTING  = "timeout"
	CMPOBJECT       = "completed"
	CMPTHRESHOLD    = "completed-threshold"
	CMPLIMIT        = "completed-limit"
	PRPLIMIT        = "prepared-limit"
	PRETTY          = "pretty"
	MAXINDEXAPI     = "max-index-api"
	PROFILE         = "profile"
	CONTROLS        = "controls"
	N1QLFEATCTRL    = "n1ql-feat-ctrl"
	AUTOPREPARE     = "auto-prepare"
)

type Checker func(interface{}) (bool, errors.Error)

var CHECKERS = map[string]Checker{
	CPUPROFILE:      checkString,
	DEBUG:           checkBool,
	KEEPALIVELENGTH: checkNumber,
	LOGLEVEL:        checkLogLevel,
	MAXPARALLELISM:  checkNumber,
	MEMPROFILE:      checkString,
	REQUESTSIZECAP:  checkNumber,
	PIPELINEBATCH:   checkNumber,
	PIPELINECAP:     checkNumber,
	SCANCAP:         checkNumber,
	SERVICERS:       checkNumber,
	TIMEOUTSETTING:  checkNumber,
	CMPOBJECT:       checkCompleted,
	CMPTHRESHOLD:    checkNumber,
	CMPLIMIT:        checkNumber,
	PRPLIMIT:        checkPositiveInteger,
	PRETTY:          checkBool,
	MAXINDEXAPI:     checkNumber,
	PROFILE:         checkProfileAdmin,
	CONTROLS:        checkControlsAdmin,
	N1QLFEATCTRL:    checkNumber,
	AUTOPREPARE:     checkBool,
}

func checkBool(val interface{}) (bool, errors.Error) {
	_, ok := val.(bool)
	return ok, nil
}

func checkNumber(val interface{}) (bool, errors.Error) {
	_, ok := val.(float64)
	return ok, nil
}

func checkObject(val interface{}) (bool, errors.Error) {
	_, ok := val.(map[string]interface{})
	return ok, nil
}

func checkCompleted(val interface{}) (bool, errors.Error) {
	object, ok := val.(map[string]interface{})
	if !ok {
		return ok, nil
	}
	for n, v := range object {
		var op RequestsOp

		switch n[0] {
		case '+':
			op = CMP_OP_ADD
			n = n[1:]
		case '-':
			op = CMP_OP_DEL
			n = n[1:]
		default:
			op = CMP_OP_UPD
		}
		err := RequestsCheckQualifier(n, op, v)
		if err != nil && op == CMP_OP_UPD && err.Code() == errors.ADMIN_QUALIFIER_NOT_SET {
			err = RequestsCheckQualifier(n, CMP_OP_ADD, v)
		}
		if err != nil {
			return false, err
		}
	}
	return ok, nil
}

func checkPositiveInteger(val interface{}) (bool, errors.Error) {
	v, ok := val.(float64)

	// we are getting floats here - val doesn't cast to ints
	// and we want a cache, however small
	return ok && (v > 1), nil
}

func checkString(val interface{}) (bool, errors.Error) {
	_, ok := val.(string)
	return ok, nil
}

func checkLogLevel(val interface{}) (bool, errors.Error) {
	level, is_string := val.(string)
	if !is_string {
		return false, nil
	}
	_, ok := logging.ParseLevel(level)
	return ok, nil
}
