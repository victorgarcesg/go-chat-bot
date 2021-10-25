package pkg

type optionID int

const (
	OPT_NICK optionID = iota
	OPT_JOIN
	OPT_ROOMS
	OPT_MSG
	OPT_QUIT
)

type Option struct {
	ID       optionID
	Client   *Client
	Argument string
}
