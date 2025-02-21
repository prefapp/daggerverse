package main

type ClaimsDataFile struct {
	RegoFile         string            `yaml:"regoFile"`
	ApplicableClaims []ApplicableClaim `yaml:"applicableClaims"`
}

type ApplicableClaim struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"`
}

type Claim struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"`
}
