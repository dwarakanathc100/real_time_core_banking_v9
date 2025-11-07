package account

import (
    "testing"
)

func TestSanity(t *testing.T) {
    // trivial test to ensure package builds
    if 1+1 != 2 { t.Fatal("math broken") }
}
