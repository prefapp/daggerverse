package main

import (
    "fmt"
)

func (m *NotifyAndHydrateState) CmdContainer(

    // +optional
    // +default="latest-slim"
    firestarterImageTag string,

    // +optional
    // +default="ghcr.io/prefapp/gitops-k8s"
    firestarterImage string,

) *Container {

    return dag.Container().

        From(fmt.Sprintf("%s:%s", firestarterImage, firestarterImageTag))

}

//func (m *NotifyAndHydrateState) Hidrate(
//
//    claimsDir *Dir
//    claimsRepo string,
//    crsDir string
//) *Container {
//
//    cmd := dag.Container().From(fmt.Sprintf("%s:%s", firestarterImage, firestarterImageTag))
//
//    return cmd
//
//
//}
