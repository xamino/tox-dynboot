package toxdynboot

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const toxWikiNodesURL = "https://wiki.tox.im/Nodes"

// ToxNode is a single possible node candidate.
type ToxNode struct {
	IPv4 string
	IPv6 string
	Port uint16
	// TODO: parse the ID to the correct format (need to dehex it I think)
	ID string
	// TODO: can I do anything with this too?
	Locale string
	Name   string
}

// Fetch the possible bootstrap nodes from the wiki. Requires active internet!
func Fetch() (*[]ToxNode, error) {
	response, err := http.Get(toxWikiNodesURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// remove stuff before first <td> and split rest up
	candidates := strings.Split(string(contents), "<td>")[1:]
	var list []string
	for _, element := range candidates {
		// clean up anything outside the </td>
		element = strings.Split(element, "</td>")[0]
		// remove any trailing newlines
		element = strings.Trim(element, "\n")
		// now each element is a single value
		list = append(list, element)
	}
	// if this changes the parsing won't work anyway, so warn and quit here
	if len(list)%7 != 0 {
		return nil, errors.New("Table is not formatted correctly, contact code maintainer of toxdynboot!")
	}
	// determine how many iterations we'll have to do
	amount := len(list) / 7
	// list of objects
	var nodes []ToxNode
	// now build Straps from the elements
	for i := 0; i < amount; i++ {
		index := i * 7
		// most we can directly assign
		object := ToxNode{IPv4: list[index], IPv6: list[index+1], ID: list[index+3], Locale: list[index+5], Name: list[index+4]}
		// port needs to converted first
		temp, err := strconv.ParseInt(list[index+2], 10, 32)
		if err != nil {
			// This means either we've made a mistake or the table is badly formatted
			return nil, err
		}
		object.Port = uint16(temp)
		// and done
		nodes = append(nodes, object)
	}
	// TODO: add liveliness check
	// return full list
	return &nodes, nil
}
