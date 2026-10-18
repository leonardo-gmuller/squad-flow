package identifier

import "testing"

func TestAPIUUIDFormat(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value string
		valid bool
	}{
		{name: "canonical", value: "a4e7fe1d-a5b2-4781-a489-c7632aab12cd", valid: true},
		{name: "uppercase remains compatible", value: "A4E7FE1D-A5B2-4781-A489-C7632AAB12CD", valid: true},
		{name: "numeric database ID", value: "123"},
		{name: "empty"},
		{name: "invalid hexadecimal", value: "z4e7fe1d-a5b2-4781-a489-c7632aab12cd"},
		{name: "unhyphenated", value: "a4e7fe1da5b24781a489c7632aab12cd"},
		{name: "URN", value: "urn:uuid:a4e7fe1d-a5b2-4781-a489-c7632aab12cd"},
		{name: "braces", value: "{a4e7fe1d-a5b2-4781-a489-c7632aab12cd}"},
		{name: "surrounding spaces", value: " a4e7fe1d-a5b2-4781-a489-c7632aab12cd "},
		{name: "invalid separators", value: "a4e7fe1d_a5b2_4781_a489_c7632aab12cd"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := IsUUID(test.value); got != test.valid {
				t.Fatalf("IsUUID(%q)=%v, want %v", test.value, got, test.valid)
			}
		})
	}
}
