package tracker

type Ticket struct {
	ID    string
	URL   string
	Title string
}

type Tracker interface {
	FetchIssue(branchName string) (*Ticket, error)
}
