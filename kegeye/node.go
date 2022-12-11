package kegeye

import (
	"errors"
	"strconv"
	"time"
)

const (
	kegTsLayout = "2006-01-02 15:04:05Z"
)

var (
	ErrNotEnoughTsvParts = errors.New("not enough tsv parts in dex entry")
)

type Node struct {
	Id        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Title     string    `json:"title"`
}

func NodeFromSlice(s []string) (*Node, error) {
	if len(s) != 3 {
		return nil, ErrNotEnoughTsvParts
	}

	id, err := strconv.Atoi(s[0])
	if err != nil {
		return nil, err
	}

	ts, err := time.Parse(kegTsLayout, s[1])
	if err != nil {
		return nil, err
	}

	node := Node{
		Id:        id,
		Timestamp: ts,
		Title:     s[2],
	}

	return &node, nil
}

func NodeSliceFromTsv(s [][]string) (NodeSlice, error) {
	sl := NodeSlice{}

	for _, v := range s {
		n, err := NodeFromSlice(v)
		if err != nil {
			return NodeSlice{}, err
		}
		sl = append(sl, *n)
	}

	return sl, nil
}

type NodeSlice []Node
type NodeMap map[int]Node

func (ns NodeSlice) ToNodeMap() NodeMap {
	m := NodeMap{}

	for _, v := range ns {
		m[v.Id] = v
	}

	return m
}
