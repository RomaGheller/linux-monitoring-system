package monitoring

type CheckLauncher interface {
	Launch(checkers []Checker) []CheckResult
}

type SequentialLauncher struct{}

func (SequentialLauncher) Launch(checkers []Checker) []CheckResult {
	results := make([]CheckResult, 0, len(checkers))

	for _, checker := range checkers {
		results = append(results, checker.Check())
	}

	return results
}
