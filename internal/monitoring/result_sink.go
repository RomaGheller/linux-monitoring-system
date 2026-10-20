package monitoring

import "fmt"

type ResultSink interface {
	Publish(results []CheckResult)
}

type StdoutResultSink struct{}

func (StdoutResultSink) Publish(results []CheckResult) {
	for _, result := range results {
		fmt.Printf(
			"%s: %s [%s] - %s\n",
			result.Name,
			result.Status,
			result.State,
			result.Message,
		)
	}
}
