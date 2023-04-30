package evaluation

type PullRequest struct {
	id          int
	title       string
	description string
	status      string
	category    string // bounty, chore, defect, etc.
	task        Task   // Link to a task could be optional
}

func NewPullRequest(id int, title string, description, status string, category string, task Task) *PullRequest {
	pr := &PullRequest{
		id:          id,
		title:       title,
		description: description,
		status:      status,
		category:    category,
		task:        task,
	}
	return pr
}

func (pr PullRequest) Id() int {
	return pr.id
}

func (pr PullRequest) Status() string {
	return pr.status
}

func (pr PullRequest) UpdateStatus(status string) {
	pr.status = status
}
