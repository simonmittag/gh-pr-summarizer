package tracker

type Ticket struct {
	ID    string
	URL   string
	Title string
}

type Tracker interface {
	FetchTicket(branchName string) (*Ticket, error)
}
