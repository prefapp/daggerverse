package main

type Version string

const (
	v1_31 Version = "v1_31"
	v1_32 Version = "v1_32"
	v1_33 Version = "v1_33"
	v1_34 Version = "v1_34"
	v1_35 Version = "v1_35"
)

var K8sVersions = map[Version]string{
	v1_35: "kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f",
	v1_34: "kindest/node:v1.34.3@sha256:08497ee19eace7b4b5348db5c6a1591d7752b164530a36f855cb0f2bdcbadd48",
	v1_33: "kindest/node:v1.33.7@sha256:d26ef333bdb2cbe9862a0f7c3803ecc7b4303d8cea8e814b481b09949d353040",
	v1_32: "kindest/node:v1.32.11@sha256:5fc52d52a7b9574015299724bd68f183702956aa4a2116ae75a63cb574b35af8",
	v1_31: "kindest/node:v1.31.14@sha256:6f86cf509dbb42767b6e79debc3f2c32e4ee01386f0489b3b2be24b0a55aac2b",
}
