package main

import (
	"testing"
)

func TestPolicyIsApplicableToClaim(t *testing.T) {

	m := &Opa{}

	claimDataFile := ClaimsDataFile{
		RegoFile: "regoFile",
		ApplicableClaims: []ApplicableClaim{
			{
				Name: "*",
				Kind: "Testkind",
			},
		},
	}

	claim := Claim{
		Name: "name",
		Kind: "Testkind",
	}

	if !m.PolicyIsApplicableToClaim(claimDataFile, claim) {

		t.Errorf("Expected true, got false")

	}

}
