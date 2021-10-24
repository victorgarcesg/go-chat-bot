package articles

type optionID int

const (
	OPT_NICK optionID = iota
	OPT_JOIN
	OPT_ROOMS
	OPT_MSG
	OPT_QUIT
)

type option struct {
	id       optionID
	client   *Client
	argument string
}
