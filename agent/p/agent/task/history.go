package task

type History struct {
	Size       int
	NumRemoved int
	Results    []Result
}

func (h *History) AddResult(result Result) {
	h.Results = append(h.Results, result)
	if numResults := len(h.Results); numResults > h.Size {
		numToRemove := numResults - h.Size
		h.Results = h.Results[numToRemove:]
		h.NumRemoved += numToRemove
	}
}
