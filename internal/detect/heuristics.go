// Copyright 2025-2026 Oliver Eikemeier. All Rights Reserved.
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

//go:generate go tool bitmask -type Heuristics

// Heuristics represents a set of heuristic flags used to control various passes in the analysis process.
type Heuristics uint8

const (
	// HeuristicVar represents a heuristic pass for variable declarations.
	HeuristicVar Heuristics = 1 << iota // var

	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage // usage

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers // receivers

	// HeuristicAll combines all available heuristic passes into a single constant, encompassing all analysis strategies.
	HeuristicAll Heuristics = 1<<iota - 1 // all

	// HeuristicOff turns off all heuristic passes.
	HeuristicOff Heuristics = 0 // off
)
