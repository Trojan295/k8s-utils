package k8siperf3

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	iperf3SenderLogRegexp = regexp.MustCompile(`sec.+\d.+ (.+) Gbits\/sec.+ \d.+ \d.+\n`)
)

func parseIperf3Logs(logs string, omitStart, omitEnd int) (*Result, error) {
	matches := iperf3SenderLogRegexp.FindAllStringSubmatch(logs, -1)
	count := len(matches) - omitStart - omitEnd

	var bitrate float64

	for _, match := range matches[omitStart : len(matches)-omitEnd] {
		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return &Result{}, fmt.Errorf("while parsing iperf3 logs: %w", err)
		}

		bitrate += value / float64(count)
	}

	return &Result{
		Bitrate: bitrate,
	}, nil
}
