package mutations

type goASTAnalyzer struct{}
type sqlDDLAnalyzer struct{}
type openAPIAnalyzer struct{}
type manifestAnalyzer struct{}
type runtimeConfigAnalyzer struct{}

func DefaultAnalyzers() []Analyzer {
	return []Analyzer{
		goASTAnalyzer{},
		sqlDDLAnalyzer{},
		openAPIAnalyzer{},
		manifestAnalyzer{},
		runtimeConfigAnalyzer{},
	}
}
