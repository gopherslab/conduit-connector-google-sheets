// // Copyright © 2022 Meroxa, Inc.
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.

package config

import (
	"testing"
)

func TestParse(t *testing.T) {
	result, err := Parse(map[string]string{
		"access_token":   "ya29.A0ARrdaM-XqU2M3QzTTr-M-WB7nLXoHYqz3YQB4bCZzIOVpsB0sE59RFUDNl-fqfhPvv4TX_h9zZOaxGtEAao9vvHtN52MmKoYtTtOWwT5ozZV3tGfqI6sJX9d3FbM17KvbXzU4qsspLXa91Kso_Ux0EzNK8ZG",
		"spreadsheet_id": "1gQjm4hnSdrMFyPjhlwSGLBbj0ACOxFQJpVST1LmW6Hg",
		"range":          "Sheet1",
	})

	if err != nil {
		t.Errorf("%v", err)
		return
	} else {
		t.Logf("Passed. %#v", result)
	}
}
