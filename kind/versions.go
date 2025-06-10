package main

type KindVersion string
type K8sVersion string

const (
	v0_25 KindVersion = "v0_25"
	v0_26 KindVersion = "v0_26"
	v0_27 KindVersion = "v0_27"
	v0_28 KindVersion = "v0_28"
	v0_29 KindVersion = "v0_29"
)

const (
	v1_26_15 K8sVersion = "v1_26_15"
	v1_27_16 K8sVersion = "v1_27_16"
	v1_28_15 K8sVersion = "v1_28_15"
	v1_29_10 K8sVersion = "v1_29_10"
	v1_29_12 K8sVersion = "v1_29_12"
	v1_29_14 K8sVersion = "v1_29_14"
	v1_30_6  K8sVersion = "v1_30_6"
	v1_30_8  K8sVersion = "v1_30_8"
	v1_30_10 K8sVersion = "v1_30_10"
	v1_30_13 K8sVersion = "v1_30_13"
	v1_31_2  K8sVersion = "v1_31_2"
	v1_31_4  K8sVersion = "v1_31_4"
	v1_31_6  K8sVersion = "v1_31_6"
	v1_31_9  K8sVersion = "v1_31_9"
	v1_32_0  K8sVersion = "v1_32_0"
	v1_32_2  K8sVersion = "v1_32_2"
	v1_32_5  K8sVersion = "v1_32_5"
	v1_33_1  K8sVersion = "v1_33_1"
)

var KindToK8s = map[KindVersion][]K8sVersion{
	v0_25: []K8sVersion{v1_26_15, v1_27_16, v1_28_15, v1_29_10, v1_30_6, v1_31_2},
	v0_26: []K8sVersion{v1_29_12, v1_30_8, v1_31_4, v1_32_0},
	v0_27: []K8sVersion{v1_29_14, v1_30_10, v1_31_6, v1_32_2},
	v0_28: []K8sVersion{v1_30_13, v1_31_9, v1_32_5, v1_33_1},
	v0_29: []K8sVersion{v1_30_13, v1_31_9, v1_32_5, v1_33_1},
}

var K8sVersions = map[K8sVersion]string{
	v1_26_15: "kindest/node:v1.26.15@sha256:c79602a44b4056d7e48dc20f7504350f1e87530fe953428b792def00bc1076dd",
	v1_27_16: "kindest/node:v1.27.16@sha256:2d21a61643eafc439905e18705b8186f3296384750a835ad7a005dceb9546d20",
	v1_28_15: "kindest/node:v1.28.15@sha256:a7c05c7ae043a0b8c818f5a06188bc2c4098f6cb59ca7d1856df00375d839251",
	v1_29_10: "kindest/node:v1.29.10@sha256:3b2d8c31753e6c8069d4fc4517264cd20e86fd36220671fb7d0a5855103aa84b",
	v1_29_12: "kindest/node:v1.29.12@sha256:62c0672ba99a4afd7396512848d6fc382906b8f33349ae68fb1dbfe549f70dec",
	v1_29_14: "kindest/node:v1.29.14@sha256:8703bd94ee24e51b778d5556ae310c6c0fa67d761fae6379c8e0bb480e6fea29",
	v1_30_6:  "kindest/node:v1.30.6@sha256:b6d08db72079ba5ae1f4a88a09025c0a904af3b52387643c285442afb05ab994",
	v1_30_8:  "kindest/node:v1.30.8@sha256:17cd608b3971338d9180b00776cb766c50d0a0b6b904ab4ff52fd3fc5c6369bf",
	v1_30_10: "kindest/node:v1.30.10@sha256:4de75d0e82481ea846c0ed1de86328d821c1e6a6a91ac37bf804e5313670e507",
	v1_30_13: "kindest/node:v1.30.13@sha256:8673291894dc400e0fb4f57243f5fdc6e355ceaa765505e0e73941aa1b6e0b80",
	v1_31_2:  "kindest/node:v1.31.2@sha256:18fbefc20a7113353c7b75b5c869d7145a6abd6269154825872dc59c1329912e",
	v1_31_4:  "kindest/node:v1.31.4@sha256:2cb39f7295fe7eafee0842b1052a599a4fb0f8bcf3f83d96c7f4864c357c6c30",
	v1_31_6:  "kindest/node:v1.31.6@sha256:28b7cbb993dfe093c76641a0c95807637213c9109b761f1d422c2400e22b8e87",
	v1_31_9:  "kindest/node:v1.31.9@sha256:156da58ab617d0cb4f56bbdb4b493f4dc89725505347a4babde9e9544888bb92",
	v1_32_0:  "kindest/node:v1.32.0@sha256:2458b423d635d7b01637cac2d6de7e1c1dca1148a2ba2e90975e214ca849e7cb",
	v1_32_2:  "kindest/node:v1.32.2@sha256:f226345927d7e348497136874b6d207e0b32cc52154ad8323129352923a3142f",
	v1_32_5:  "kindest/node:v1.32.5@sha256:36187f6c542fa9b78d2d499de4c857249c5a0ac8cc2241bef2ccd92729a7a259",
	v1_33_1:  "kindest/node:v1.33.1@sha256:8d866994839cd096b3590681c55a6fa4a071fdaf33be7b9660e5697d2ed13002",
}
