package toxdynboot

import "errors"

/*
Errors used.
*/
var (
	errSourceFormat = errors.New("source can not be parsed")
	errSourceTable  = errors.New("source table not parseable")
	errAliveTimeout = errors.New("alive timed out")
	errNoToxNodes   = errors.New("no ToxNodes could be fetched")
)

/*
toxWikiNodesURL is the address where we look for the node list. Note that we use
the ?do=edit to parse the raw markup instead of the html (which keeps changing).
*/
const toxWikiNodesURL = "https://wiki.tox.chat/users/nodes?do=edit"

/*
tableColumns is the amount of columns we expect the table to have to be parseable.
*/
const tableColumns = 6

/*
Specifies in which column which value is stored. Amount of specified columns
should match tableColumns and have a mapping in ToxNode.
*/
const (
	columnIP4        = 0
	columnIP6        = 1
	columnPort       = 2
	columnKey        = 3
	columnMaintainer = 4
	columnLocation   = 5
)
