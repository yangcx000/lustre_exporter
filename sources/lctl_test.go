// (C) Copyright 2021 Gabriele Iannetti <g.iannetti@gsi.de>
// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
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

package sources

import (
	"testing"
)

func TestChangelogTarget(t *testing.T) {
	testBlock := `mdd.lustre-MDT0000.changelog_users=`
	expected := "lustre-MDT0000"

	result, err := regexCaptureChangelogTarget(testBlock)

	if err != nil {
		t.Fatal(err)
	}

	if expected != result {
		t.Fatalf("No changelog target found. Expected: %s, Got %s", expected, result)
	}

	testBlock = `mdd..changelog_users=`

	_, err = regexCaptureChangelogTarget(testBlock)

	if err == nil {
		t.Fatal("Expected error if not changelog target has been found")
	}
}

func TestChangelogCurrentIndex(t *testing.T) {
	testBlock := `mdd.lustre-MDT0000.changelog_users=
	current index: 34
	ID    index (idle seconds)
	cl1   0 (1725676)
	cl2   34 (28)`
	expected := float64(34)

	result, err := regexCaptureChangelogCurrentIndex(testBlock)

	if err != nil {
		t.Fatal(err)
	}

	if expected != result {
		t.Fatalf("Retrieved an unexpected value. Expected: %f, Got %f", expected, result)
	}

	testBlock = `mdd.lustre-MDT0000.changelog_users=
	current index: 0`
	expected = 0

	result, err = regexCaptureChangelogCurrentIndex(testBlock)

	if err != nil {
		t.Fatal(err)
	}

	if expected != result {
		t.Fatalf("Retrieved an unexpected value. Expected: %f, Got %f", expected, result)
	}

	testBlock = `mdd.lustre-MDT0000.changelog_users=
	ID    index (idle seconds)`

	_, err = regexCaptureChangelogCurrentIndex(testBlock)

	if err == nil {
		t.Fatal("Expected error if no current changelog index has been found")
	}
}

func TestChangelogUser(t *testing.T) {
	testBlock := `mdd.lustre-MDT0000.changelog_users=
	current index: 34
	ID    index (idle seconds)
	cl1   0 (1725676)
	cl2   34 (28)`

	result := regexCaptureChangelogUser(testBlock)

	if len(result) != 2 {
		t.Fatalf("Retrieved unexpected length of changelog reader. Expected: %d, Got: %d", 2, len(result))
	}

	expected := "cl1   0 (1725676)"
	matched := result[0][0]

	if expected != matched {
		t.Fatalf("Retrieved an unexpected value. Expected: %s, Got: %s", expected, matched)
	}

	expected = "cl2   34 (28)"
	matched = result[1][0]

	if expected != matched {
		t.Fatalf("Retrieved an unexpected value. Expected: %s, Got: %s", expected, matched)
	}
}
