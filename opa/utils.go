package main

import "regexp"

func strip(str string) string {

	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	var re = regexp.MustCompile(ansi)

	str = re.ReplaceAllString(str, "")

	str = regexp.MustCompile(`\n`).ReplaceAllString(str, "")

	// remove carriage return
	str = regexp.MustCompile(`\r`).ReplaceAllString(str, "")

	// remove line feed
	str = regexp.MustCompile(`\f`).ReplaceAllString(str, "")

	return str
}
