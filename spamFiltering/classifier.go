package spamFiltering

import (
	"github.com/jbrukh/bayesian"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

func Train() (c *bayesian.Classifier) {
	classifier := bayesian.NewClassifier(Good, Bad)
	goodStuff := []string{"manga", "artists", "phore"}
	badStuff := []string{"porn", "child porn", "hitman", "missiles", "human trafficking", "Manga"}
	classifier.Learn(goodStuff, Good)
	classifier.Learn(badStuff, Bad)
	classifier.ConvertTermsFreqToTfIdf()
	return classifier
}
