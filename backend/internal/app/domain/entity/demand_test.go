package entity

import "testing"

func TestAllowedTransitions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		from, to Status
		want     bool
	}{
		{Backlog, InProgress, true},
		{Backlog, Done, false},
		{InProgress, InQA, true},
		{InQA, Done, true},
		{InQA, ChangesRequested, true},
		{ChangesRequested, InProgress, true},
		{Done, InProgress, false},
		{InQA, InQA, false},
	}
	for _, tc := range cases {
		if got := CanTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("%s -> %s = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()

	item := Demand{Title: "Hello world", Kind: Improvement, Priority: "high"}
	if err := item.Validate(); err != nil {
		t.Fatal(err)
	}

	item.Kind = "invalid code!"
	if err := item.Validate(); err == nil {
		t.Fatal("invalid kind accepted")
	}

	tc := TestCase{Title: "Breaks rule", Result: "failed"}
	if err := tc.Validate(); err == nil {
		t.Fatal("failed test without actual result accepted")
	}
}
