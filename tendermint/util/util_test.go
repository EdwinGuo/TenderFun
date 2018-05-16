package util

import "testing"

func TestingUtil(t *testing.T) {
    lookup1 := ParseCity("/Users/edwinguo/go/src/tenderfun/tendermint/test/testcase1.csv")
    if len(lookup1) != 12 {
       t.Errorf("parseCity was incorrect for testcase1, got: %d, want: %d.", len(lookup1), 12)
    }

    lookup2 := ParseCity("/Users/edwinguo/go/src/tenderfun/tendermint/test/testcase2.csv")
    if len(lookup2) != 12 {
       t.Errorf("parseCity was incorrect for testcase2, got: %d, want: %d.", len(lookup2), 12)
    }

}
