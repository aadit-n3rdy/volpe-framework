package types

type Individual struct {
	Genotype []byte
	Fitness  float32
}

type Population struct {
	ProblemID   string
	Individuals []Individual
}
