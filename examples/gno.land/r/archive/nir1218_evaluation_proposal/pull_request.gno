package evaluation

var pullRequestStatus map[string]struct{}

type PullRequest struct {
	id          int
	name        string
	description string
	status      string // Draft, Review required, Changes requested, Approved
	category    string // bounty, chore, defect, document etc.
}

func init() {
	pullRequestStatus = make(map[string]struct{})
	pullRequestStatus["Draft"] = struct{}{}
	pullRequestStatus["Approved"] = struct{}{}
	pullRequestStatus["Changes requested"] = struct{}{}
	pullRequestStatus["Review required"] = struct{}{}
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
	_, ok := pullRequestStatus[status]
	return ok
}
