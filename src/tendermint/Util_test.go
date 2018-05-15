package main

import "testing"

func TestingUtil(t *testing.T) {
    lookup1 := parseCity("/Users/edwinguo/edwin/TenderFun/src/test/testcase1.csv")
    if len(lookup1) != 12 {
       t.Errorf("parseCity was incorrect for testcase1, got: %d, want: %d.", len(lookup1), 12)
    }

    lookup2 := parseCity("/Users/edwinguo/edwin/TenderFun/src/test/testcase2.csv")
    if len(lookup2) != 12 {
       t.Errorf("parseCity was incorrect for testcase2, got: %d, want: %d.", len(lookup2), 12)
    }

}
