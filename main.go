/*
Package toxdynboot requests the Node list from the Tox wiki and provides helpful
functionality for them. Probable way to use this package is via the FetchAnyAlive()
function.
*/
package toxdynboot

import (
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
FetchAll returns all nodes that are in the wiki.
*/
func FetchAll() ([]ToxNode, error) {
	return parseNodes()
}

/*
FetchAny returns a random single node from the wiki.
*/
func FetchAny() (*ToxNode, error) {
	nodesTemp, err := parseNodes()
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
	nodes, err := parseNodes()
	if err != nil {
		return nil, err
	}
	c := make(chan *ToxNode)
	for _, node := range nodes {
		// concurrently do this because it locks.
		go func(node ToxNode) {
			if isAlive(&node, timeout) {
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
within the given timeout. This is the method you should probably use to bootstrap
a client with multiple Tox nodes.
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
FetchFirstAlive will return the first node that we determine to be available. The timeout
is the max time: if reached the function will return an error.
*/
func FetchFirstAlive(timeout time.Duration) (*ToxNode, error) {
	// we'll only check those marked as active
	nodes, err := parseNodes()
	if err != nil {
		return nil, err
	}
	// prevent freezing if no nodes were fetched
	if len(nodes) == 0 {
		return nil, nil
	}
	c := make(chan *ToxNode)
	for _, node := range nodes {
		// concurrently do this because it locks.
		go func(node ToxNode) {
			if isAlive(&node, timeout) {
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
	return nil, nil
}

/*
parseNodes reads the possible bootstrap nodes from the wiki. Requires active internet!
*/
func parseNodes() ([]ToxNode, error) {
	// TODO: this can block for a long time – implement timeout?
	response, err := http.Get(toxWikiNodesURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	strContents := string(contents)
	// check if we can locate the source text with the table
	if !strings.Contains(strContents, "|") {
		return nil, errSourceFormat
	}
	// so long as only a single table is on that wiki page, this split will
	// return all entry candidates
	splitted := strings.Split(strContents, "|")
	// if we have no candidates we're done, probabaly formatting error
	if len(splitted) == 0 {
		return nil, errSourceFormat
	}
	// throw away first entry (contain start of webpage)
	splitted = splitted[1 : len(splitted)-2] // TODO note -2: why?
	// remove empty elements
	var list []string
	for _, cand := range splitted {
		// trim space
		cand = strings.TrimSpace(cand)
		// if trim results in empty candidate, throw away (can be line break)
		if cand == "" {
			continue
		}
		// append remaining
		list = append(list, cand)
	}
	// if this changes the parsing won't work anyway, so warn and quit here
	if len(list)%tableColumns != 0 {
		return nil, errSourceTable
	}
	// determine how many iterations we'll have to do
	amount := len(list) / tableColumns
	// list of objects
	var nodes []ToxNode
	// now build ToxNodes from the elements
	for i := 0; i < amount; i++ {
		index := i * tableColumns
		// most we can directly assign
		object := ToxNode{
			IPv4:       list[index+columnIP4],
			IPv6:       list[index+columnIP6],
			Maintainer: list[index+columnMaintainer],
			Location:   list[index+columnLocation]}
		// port needs to converted first
		temp, err := strconv.ParseInt(list[index+columnPort], 10, 32)
		if err != nil {
			// This means either we've made a mistake or the table is badly formatted
			return nil, err
		}
		// needs to be cast to correct value too
		object.Port = uint16(temp)
		// key needs to be dehexed
		object.PublicKey, err = hex.DecodeString(list[index+columnKey])
		if err != nil {
			return nil, err
		}
		// and done
		nodes = append(nodes, object)
	}
	// return full list
	return nodes, nil
}

/*
IsAlive checks whether the given ToxNode is reachable. NOTE: this relies on nodes refusing
connections - if they are online but quietly discard connection attempts, this function
will wrongly label them as unreachable (which is the case for a few of the current nodes
as of 2015.06.10).
*/
func isAlive(node *ToxNode, timeout time.Duration) bool {
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
