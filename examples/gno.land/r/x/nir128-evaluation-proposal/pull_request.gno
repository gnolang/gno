package evaluation

var statuses map[string]struct{}

type PullRequest struct {
	id          int
	name        string
	description string
	status      string // Draft, Review required, Changes requested, Approved
	category    string // bounty, chore, defect, etc.
}

func init() {
	statuses = make(map[string]struct{})
	statuses["Draft"] = struct{}{}
	statuses["Approved"] = struct{}{}
	statuses["Changes requested"] = struct{}{}
	statuses["Review required"] = struct{}{}
}

func NewPullRequest(id int, name string, description string, status string, category string) *PullRequest {
	pr := &PullRequest{
		id:          id,
		name:        name,
		description: description,
		status:      status,
		category:    category,
	}
	return pr
}

func (pr PullRequest) Id() int {
	return pr.id
}

func (pr PullRequest) Status() string {
	return pr.status
}

func (pr *PullRequest) UpdateName(name string) {
	pr.name = name
}

func (pr *PullRequest) UpdateDescription(description string) {
	pr.description = description
}

func (pr *PullRequest) UpdateStatus(status string) bool {
	if validateStatus(status) {
		pr.status = status
		return true
	}
	return false
}

func validateStatus(status string) bool {
	_, ok := statuses[status]
	return ok
}
