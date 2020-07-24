/*
Copyright 2018 Aspen Mesh Authors.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the
License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
*/

package mtlspolicyutil

// State of the mTLS settings
type MTLSSetting int32

const (
	// Unknown if state cannot be determined
	MTLSSetting_UNKNOWN MTLSSetting = 0
	// Enabled if mTLS is turned on
	MTLSSetting_ENABLED MTLSSetting = 1
	// Disabled if mTLS is turned off
	MTLSSetting_DISABLED MTLSSetting = 2
	// Mixed if mTLS is partially enabled or disabled
	MTLSSetting_MIXED MTLSSetting = 3
)

// getMTLSBool returns a bool and error from the 4 possible enum mTls states.
// Mixed counts as enabled since it allows enabled traffic, but it returns an
// error in case the caller needs to know if the true status means it's
// enabled-only, or enabled in a way that allows other traffic. Unknown counts
// as disabled since we cannot tell the caller that the status is mTls enabled.
// It returns an error in case the caller needs to know if the false status
// means that the false status is actually bogus because we we unable to
// determine the mTls status.
func getMTLSBool(mtlsState MTLSSetting) bool {
	// pass in the policy to maintain the structure of returns for callers
	// pre-Oct2018-refactor.
	switch checkState := mtlsState; {
	case checkState == MTLSSetting_ENABLED:
		return true
	case checkState == MTLSSetting_UNKNOWN:
		return false
	case checkState == MTLSSetting_MIXED:
		return true
	default:
		return false
	}
}
