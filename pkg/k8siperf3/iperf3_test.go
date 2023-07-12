package k8siperf3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIperf3Logs(t *testing.T) {
	output := `Connecting to host 100.64.0.7, port 5201
[  5] local 100.64.0.165 port 46716 connected to 100.64.0.7 port 5201
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec  1.34 GBytes  11.5 Gbits/sec    0   1.82 MBytes       
[  5]   1.00-2.00   sec  1.25 GBytes  10.7 Gbits/sec    0   2.75 MBytes       
[  5]   2.00-3.00   sec  1.19 GBytes  10.3 Gbits/sec    0   3.04 MBytes       
[  5]   3.00-4.00   sec  1.31 GBytes  11.2 Gbits/sec    0   3.04 MBytes       
[  5]   4.00-5.00   sec  1.22 GBytes  10.5 Gbits/sec    0   3.04 MBytes       
[  5]   5.00-6.00   sec  1.17 GBytes  10.0 Gbits/sec    0   3.04 MBytes       
[  5]   6.00-7.00   sec  1.49 GBytes  12.8 Gbits/sec    0   3.04 MBytes       
[  5]   7.00-8.00   sec  1.67 GBytes  14.3 Gbits/sec    0   3.04 MBytes       
[  5]   8.00-9.00   sec  1.35 GBytes  11.6 Gbits/sec    0   3.04 MBytes       
[  5]   9.00-10.00  sec  1.24 GBytes  10.6 Gbits/sec    0   3.04 MBytes       
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  13.2 GBytes  11.4 Gbits/sec    0             sender
[  5]   0.00-10.00  sec  13.2 GBytes  11.4 Gbits/sec                  receiver

iperf Done.
`

	result, err := parseIperf3Logs(output, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 11.4, result.Bitrate)
}
