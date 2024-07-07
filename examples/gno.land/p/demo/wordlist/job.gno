package wordlist

import "math/rand"

var (
	jobs = []string{
		"architect",
		"baker",
		"cleaner",
		"cooker",
		"doctor",
		"designer",
		"developer",
		"engineer",
		"officer",
		"priest",
		"teacher",
	}
	jobsL = len(jobs)
)

func RandomJob(r *rand.Rand) string {
	return jobs[r.IntN(jobsL)]
}
