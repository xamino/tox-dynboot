/*
Package toxdynboot requests the Node list from the Tox wiki and provides helpful
functionality for them. Fastest and best way to use this package is via the
FetchFirstAlive() function.
*/
package toxdynboot

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// toxWikiNodesURL is the address where we look for the node list.
const toxWikiNodesURL = "https://wiki.tox.chat/users/nodes"

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

func (t *ToxNode) String() string {
	return "ToxNode " + t.Name + " at " + t.IPv4 + ":" + strconv.FormatInt(int64(t.Port), 10) + "."
}

/*
FetchFirstAlive will return the first node that we determine to be available. The timeout
is the max time: if reached the function will return an error.
*/
func FetchFirstAlive(timeout time.Duration) (*ToxNode, error) {
	// we'll only check those marked as active
	nodes, err := FetchUp()
	if err != nil {
		return nil, err
	}
	// prevent freezing if no nodes were fetched
	if len(nodes) == 0 {
		return nil, errors.New("no nodes could be fetched from URL")
	}
	c := make(chan *ToxNode)
	for _, node := range nodes {
		// concurrently do this because it locks.
		go func(node ToxNode) {
			if IsAlive(&node, timeout) {
				c <- &node
			} else {
				c <- nil
			}
		}(node)
		// warning: don't use node directly in the anonymous function because it changes on every iteration!
	}
	candidate := <-c
	if candidate != nil {
		return candidate, nil
	}
	return nil, errors.New("No ToxNode could be reached!")
}

/*
FetchAll the possible bootstrap nodes from the wiki. Requires active internet!
*/
func FetchAll() ([]ToxNode, error) {
	response, err := http.Get(toxWikiNodesURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// locate only the table we are interested in
	trimmed := strings.Split(string(contents), "id=\"active_nodes_list\"")[1]
	trimmed = strings.Split(trimmed, "id=\"running_a_node\"")[0]
	// remove stuff before first <td> and split rest up
	candidates := strings.Split(trimmed, "<td")[1:]
	// if we have no candidates we're done, probabaly formatting error
	if len(candidates) == 0 {
		return nil, errors.New("parsing of wiki table failed")
	}
	var list []string
	for _, element := range candidates {
		// clean up leading stuff
		element = strings.Split(element, ">")[1]
		// clean up anything outside the </td>
		element = strings.Split(element, "</td")[0]
		// remove whitespace
		element = strings.TrimSpace(element)
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
	// now build ToxNodes from the elements
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
	return nodes, nil
}

/*
FetchUp returns all nodes that are marked as being online in the wiki.
*/
func FetchUp() ([]ToxNode, error) {
	nodes, err := FetchAll()
	if err != nil {
		return nil, err
	}
	var upNodes []ToxNode
	for _, node := range nodes {
		if node.Status {
			upNodes = append(upNodes, node)
		}
	}
	return upNodes, nil
}

/*
FetchAny returns a random single node with the status of UP from the wiki.
*/
func FetchAny() (*ToxNode, error) {
	nodesTemp, err := FetchUp()
	if err != nil {
		return nil, err
	}
	// shortcut
	if len(nodesTemp) == 0 {
		return nil, nil
	}
	nodes := nodesTemp
	// random seed based on time (doesn't need to be cryptographically secure)
	rand.Seed(time.Now().UnixNano())
	// pick one random
	node := nodes[rand.Intn(len(nodes))]
	return &node, nil
}

/*
FetchAlive fetches all nodes from the wiki and then checks whether they are actively
reachable and only returns those. Note that this means that this function will block for
the specified time!
*/
func FetchAlive(timeout time.Duration) ([]ToxNode, error) {
	// we'll only check those marked as active
	nodes, err := FetchUp()
	if err != nil {
		return nil, err
	}
	c := make(chan *ToxNode)
	for _, node := range nodes {
		// concurrently do this because it locks.
		go func(node ToxNode) {
			if IsAlive(&node, timeout) {
				c <- &node
			} else {
				c <- nil
			}
		}(node)
		// warning: don't use node directly in the anonymous function because it changes on every iteration!
	}
	var aliveNodes []ToxNode
	for i := 0; i < len(nodes); i++ {
		candidate := <-c
		if candidate != nil {
			aliveNodes = append(aliveNodes, *candidate)
		}
	}
	// log.Printf("Of %2d nodes %2d are alive.", len(*nodes), len(aliveNodes))
	return aliveNodes, nil
}

/*
FetchAnyAlive will retrive a random node of those that have been determined to be alive
within the given timeout. NOTE: Usually you should use FetchFirstAlive() instead of this
function.
*/
func FetchAnyAlive(timeout time.Duration) (*ToxNode, error) {
	nodesTemp, err := FetchAlive(timeout)
	if err != nil {
		return nil, err
	}
	// shortcut
	if len(nodesTemp) == 0 {
		return nil, nil
	}
	nodes := nodesTemp
	// random seed based on time (doesn't need to be cryptographically secure)
	rand.Seed(time.Now().UnixNano())
	// pick one random
	node := nodes[rand.Intn(len(nodes))]
	return &node, nil
}

/*
IsAlive checks whether the given ToxNode is reachable. NOTE: this relies on nodes refusing
connections - if they are online but quietly discard connection attempts, this function
will wrongly label them as unreachable (which is the case for a few of the current nodes
as of 2015.06.10).
*/
func IsAlive(node *ToxNode, timeout time.Duration) bool {
	// TODO: use both IPv4 AND IPv6.
	address := node.IPv4 + ":" + strconv.FormatUint(uint64(node.Port), 10)
	// since ICMP ping is not trivially available we rely on the servers denying TCP connections as a ping
	conn, err := net.DialTimeout("tcp", address, timeout)
	// if err but not 'connection refused' --> unreachable for ping
	if err != nil && !strings.Contains(err.Error(), "connection refused") {
		// log.Printf("Node %s is unreachable!", node.IPv4)
		return false
	} // else if conn ok or conn refused --> alive
	// if conn happened make sure to close it as we don't need it
	if conn != nil {
		conn.Close()
	}
	return true
}
