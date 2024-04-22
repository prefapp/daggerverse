package main

import (
    "fmt"
)

func (m *NotifyAndHydrateState) CmdContainer() *Container {

    return dag.Container().

        From(fmt.Sprintf("%s:%s", m.FirestarterImage, m.FirestarterImageTag))

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
