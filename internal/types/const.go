// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package types

const (
	BuildNumberWindowsServer2025 uint32 = 26100
	BuildNumberWindowsServer2022 uint32 = 20348
	BuildNumberWindowsServer2019 uint32 = 17763
	BuildNumberWindowsServer2016 uint32 = 14393

	Namespace = "windows"
)
