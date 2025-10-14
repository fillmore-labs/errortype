// Copyright 2025 Oliver Eikemeier. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package detect

// Heuristic represents heuristic flags used to control various passes in the analysis process.
type Heuristic uint8

//go:generate go tool stringer -type Heuristic -linecomment
const (
	// HeuristicOff turns off all heuristic passes.
	HeuristicOff Heuristic = iota // off

	// HeuristicVar represents a heuristic pass for variable declarations.
	HeuristicVar // var

	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage // usage

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers // receivers
)
