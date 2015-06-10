package toxdynboot

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const toxWikiNodesURL = "https://wiki.tox.im/Nodes"

// ToxNode is a single possible node candidate.
type ToxNode struct {
	IPv4      string
	IPv6      string
	Port      uint16
	PublicKey []byte
	// TODO: can I do anything with this too?
	Locale string
	Name   string
	Status bool
}

// FetchAll the possible bootstrap nodes from the wiki. Requires active internet!
func FetchAll() (*[]ToxNode, error) {
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
		object := ToxNode{IPv4: list[index], IPv6: list[index+1], Locale: list[index+5], Name: list[index+4]}
		// port needs to converted first
		temp, err := strconv.ParseInt(list[index+2], 10, 32)
		if err != nil {
			// This means either we've made a mistake or the table is badly formatted
			return nil, err
		}
		// needs to be cast to correct value too
		object.Port = uint16(temp)
		// key needs to be dehexed
		object.PublicKey, err = hex.DecodeString(list[index+3])
		if err != nil {
			return nil, err
		}
		// status should be set
		object.Status = strings.Contains(list[index+6], "UP")
		// and done
		nodes = append(nodes, object)
	}
	// return full list
	return &nodes, nil
}

// FetchUp returns all nodes that are marked as being online in the wiki.
func FetchUp() (*[]ToxNode, error) {
	nodes, err := FetchAll()
	if err != nil {
		return nil, err
	}
	var upNodes []ToxNode
	for _, node := range *nodes {
		if node.Status {
			upNodes = append(upNodes, node)
		}
	}
	return &upNodes, nil
}

// FetchAny returns a random single node with the status of UP from the wiki.
func FetchAny() (*ToxNode, error) {
	nodesTemp, err := FetchUp()
	if err != nil {
		return nil, err
	}
	nodes := *nodesTemp
	// random seed based on time (doesn't need to be cryptographically secure)
	rand.Seed(time.Now().UnixNano())
	// pick one random
	node := nodes[rand.Intn(len(nodes))]
	return &node, nil
}
