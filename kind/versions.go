package main

type Version string

const (
	v1_26 Version = "v1_26"
	v1_27 Version = "v1_27"
	v1_28 Version = "v1_28"
	v1_29 Version = "v1_29"
	v1_30 Version = "v1_30"
	v1_31 Version = "v1_31"
	v1_32 Version = "v1_32"
)

var Versions = map[Version]string{
	v1_26: "kindest/node:v1.26.15@sha256:c79602a44b4056d7e48dc20f7504350f1e87530fe953428b792def00bc1076dd",
	v1_27: "kindest/node:v1.27.16@sha256:2d21a61643eafc439905e18705b8186f3296384750a835ad7a005dceb9546d20",
	v1_28: "kindest/node:v1.28.15@sha256:a7c05c7ae043a0b8c818f5a06188bc2c4098f6cb59ca7d1856df00375d839251",
	v1_29: "kindest/node:v1.29.10@sha256:3b2d8c31753e6c8069d4fc4517264cd20e86fd36220671fb7d0a5855103aa84b",
	v1_30: "kindest/node:v1.30.6@sha256:b6d08db72079ba5ae1f4a88a09025c0a904af3b52387643c285442afb05ab994",
	v1_31: "kindest/node:v1.31.2@sha256:18fbefc20a7113353c7b75b5c869d7145a6abd6269154825872dc59c1329912e",
	v1_32: "kindest/node:v1.32.0@sha256:2458b423d635d7b01637cac2d6de7e1c1dca1148a2ba2e90975e214ca849e7cb",
}
