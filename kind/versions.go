package main

type K8sVersion string

const (
	v1_30_13 K8sVersion = "v1_30_13"
	v1_31_9  K8sVersion = "v1_31_9"
	v1_32_5  K8sVersion = "v1_32_5"
	v1_33_1  K8sVersion = "v1_33_1"
)

var K8sVersions = map[K8sVersion]string{
	v1_30_13: "kindest/node:v1.30.13@sha256:8673291894dc400e0fb4f57243f5fdc6e355ceaa765505e0e73941aa1b6e0b80",
	v1_31_9:  "kindest/node:v1.31.9@sha256:156da58ab617d0cb4f56bbdb4b493f4dc89725505347a4babde9e9544888bb92",
	v1_32_5:  "kindest/node:v1.32.5@sha256:36187f6c542fa9b78d2d499de4c857249c5a0ac8cc2241bef2ccd92729a7a259",
	v1_33_1:  "kindest/node:v1.33.1@sha256:8d866994839cd096b3590681c55a6fa4a071fdaf33be7b9660e5697d2ed13002",
}
