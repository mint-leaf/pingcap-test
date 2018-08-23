package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var iterators = make([]*iterator, 0)

//iterator array iterator
type iterator struct {
	V    string
	Next *iterator
	I    int
}

type carrier struct {
	V *iterator
	W int
	L int
}

func (i *iterator) Index() int {
	if i.I != 0 {
		return i.I
	}
	x := strings.Index(i.V, "(")
	i.I, _ = strconv.Atoi(i.V[0:x])
	return i.I
}

func runParallel() {
	readData()
	a := time.Now().UnixNano()
	if len(iterators)%2 == 1 {
		iterators = append(iterators, nil)
	}
	result := dispatchCarrier(iterators)
	fmt.Printf("time: %d\n", time.Now().UnixNano()-a)
	count := 0
	for ; result.Next != nil; result = result.Next {
		count++
	}
	fmt.Printf("%d\n", count)
}

func main() {
	runParallel()
}

func readData() {
	for i := 0; i < 8; i++ {
		file, err := os.OpenFile(fmt.Sprintf("a-%d.json", i), os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			log.Fatalln(err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalln(err)
		}
		elements := make([]string, 0)
		err = jsoniter.Unmarshal(data, &elements)
		if err != nil {
			log.Fatalln(err)
		}
		root := new(iterator)
		c := root
		for index := range elements {
			c.Next = &iterator{
				V:    elements[index],
				Next: nil,
			}
			c = c.Next
		}
		iterators = append(iterators, root.Next)
	}
}

//combine combine two iterator
func combine(i, j *iterator) *iterator {
	temp := new(iterator)
	var m, n *iterator
	c := temp
	c.I = -1
	for m, n = i, j; m != nil && n != nil; {
		if m.Index() <= n.Index() { // append c
			c.Next = m
			m = m.Next
			c = c.Next
		} else if n.Index() != c.Index() { // append c
			c.Next = n
			n = n.Next
			c = c.Next
		} else {
			n = n.Next //not append c
		}
	}
	if m != nil {
		c.Next = m
	} else if n != nil {
		c.Next = n
	}
	return temp.Next
}

func dispatchCarrier(iters []*iterator) *iterator {
	channel := make(chan *carrier)
	for len(iters) > 1 {
		for index := range iters {
			if index%2 == 0 {
				go func(channel *chan *carrier, i, j *iterator, m, n int) {
					*channel <- calculate(&carrier{
						V: i,
						W: m,
					}, &carrier{
						V: j,
						W: n,
					})
				}(&channel, iters[index], iters[index+1], index, index+1)
			} else {
				continue
			}
		}
		other := make([]*iterator, len(iters)/2, len(iters)/2)
		//TODO:这里还可以进行优化，理论上是可以不用进行分配内存的，
		//靠多个channel进行调度，但是能力有限，目前完成不了，希望有大佬告知能怎么做
		for i := 0; i < len(iters)/2; i++ {
			other[i] = (<-channel).V
		}
		if len(other)%2 == 1 && len(other) > 1 {
			other = append(other, nil)
		}
		iters = other
	}
	return iters[0]
}

func calculate(i, j *carrier) *carrier {
	if i.W < j.W {
		return &carrier{
			V: combine(i.V, j.V),
			W: i.W,
		}
	}
	return &carrier{
		V: combine(j.V, i.V),
		W: j.W,
	}
}
