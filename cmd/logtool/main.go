package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	rex    = regexp.MustCompile("(\\S+)=(\".*?\"|\\S+)")
	layout = "2006-01-02T15:04:05.999999999Z"
)

func main() {

	fi, _ := os.Stdin.Stat()
	isPipe := false
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		isPipe = true
	}

	showQuery := flag.Bool("query", false, "show the query")
	minDur := flag.Duration("dur", 0, "only show queries which took longer than this duration, e.g. 10s")
	utc := flag.Bool("utc", false, "show timestamp in UTC time")
	flag.Parse()

	// use tabwriter to format the output spaced out
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// headings
	if *showQuery {
		fmt.Fprintf(w, "Timestamp\tTraceID\tLength\tDuration\tStatus\tPath\tQuery\n")
	} else {
		fmt.Fprintf(w, "Timestamp\tTraceID\tLength\tDuration\tStatus\tPath\n")
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		show := false
		var qt time.Time
		var dur, len time.Duration
		var status, path, query, trace string
		var err error

		line := scanner.Text()
		data := rex.FindAllStringSubmatch(line, -1)

		for _, d := range data {
			switch d[1] {
			case "traceID":
				trace = d[2]
			case "ts":
				qt, err = time.Parse(layout, d[2])
				if err != nil {
					fmt.Println(err, line)
					continue
				}
			case "msg":
				msg := strings.ReplaceAll(d[2], "\"", "")
				if !(strings.HasPrefix(msg, "GET /loki/api/") ||
					strings.HasPrefix(msg, "GET /api/prom")) {
					continue
				}

				show = true

				parts := strings.Split(msg, " ")
				u, err := url.Parse(parts[1])
				if err != nil {
					fmt.Println(err, line)
					continue
				}
				vals, err := url.ParseQuery(u.RawQuery)
				if err != nil {
					fmt.Println(err, line)
					continue
				}
				start := vals.Get("start")
				end := vals.Get("end")

				if start != "" && end != "" {
					var st, et int64
					var stErr, etErr error
					// First try to parse as unix sec timestamps
					st, stErr = strconv.ParseInt(start, 10, 64)
					et, etErr = strconv.ParseInt(end, 10, 64)
					if stErr != nil || etErr != nil {
						//Next try to parse as time/date
						var st, et time.Time
						st, stErr = time.Parse(layout, start)
						et, etErr = time.Parse(layout, end)
						if stErr != nil || etErr != nil {
							fmt.Println(err, line)
							continue
						}
						len = et.Sub(st)
					} else {
						// Loki queries are nanosecond, simple check to see if it's a second or nanosecond timestamp
						if st > 9999999999 {
							len = time.Unix(0, et).Sub(time.Unix(0, st))
						} else {
							len = time.Unix(et, 0).Sub(time.Unix(st, 0))
						}

					}

				}

				status = parts[2]
				path = u.Path
				query = vals.Get("query")

				dur, err = time.ParseDuration(parts[3])
				if err != nil {
					fmt.Println(err, line)
					continue
				}
			}
		}

		if show && dur > *minDur {
			var ts string
			if *utc {
				ts = fmt.Sprint(qt)
			} else {
				ts = fmt.Sprint(qt.Local())
			}
			if *showQuery {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n", ts, trace, len, dur, status, path, query)
			} else {
				fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", ts, trace, len, dur, status, path)
			}
			//If looking at stdin, flush after every line as someone would only paste one line at a time at the terminal
			if !isPipe {
				w.Flush()
			}

		}
		show = false
	}

	fmt.Printf("\n")
	w.Flush()

}
