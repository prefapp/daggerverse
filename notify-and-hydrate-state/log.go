package main

import "fmt"

func logPrs(message string, prs []Pr) {

	for _, pr := range prs {

		fsLog(fmt.Sprintf("%s: %v", message, pr))

	}

}

func fsLog(message string) {

	fmt.Printf("firestartr > %s\n", message)

}
