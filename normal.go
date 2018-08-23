package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/json-iterator/go"
)

var timer = make(chan int64)

func generateExamples() {
	for i := 0; i < 8; i++ {
		a := make([]string, 0)
		b := make(map[int]string)
		c := make([]int, 0)
		rand.Seed(time.Now().UnixNano())
		for {
			if len(b) == 100000 {
				break
			}
			gen := rand.Intn(1000000)
			if b[gen] != "" {
				continue
			}
			b[gen] = fmt.Sprintf("%d(%d)", gen, rand.Intn(100000))
			c = append(c, gen)
		}
		sort.Ints(c)
		for index := range c {
			a = append(a, b[c[index]])
		}
		file, err := os.OpenFile(fmt.Sprintf("a-%d.json", i), os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			log.Fatalln(err)
		}
		data, _ := jsoniter.Marshal(a)
		file.Write(data)
		time.Sleep(time.Second)
	}
	fmt.Println("generate successfully")
}

// func main() {
// 	runNormal()
// }

//RunNormal run normal code
func runNormal() {
	data := make(chan *[][]string)
	go func() {
		data <- readFile()
	}()
	a := <-timer
	length := normal(<-data)
	fmt.Printf("time: %d\n", time.Now().UnixNano()-a)
	fmt.Printf("length: %d\n", length)
}

func normal(iterators *[][]string) int {
	channel := make(chan string)
	go getDataNormal(iterators, &channel)
	v := make([]string, 1000000, 1000000)
	for {
		element := <-channel
		if element == "" {
			break
		}
		i := strings.Index(element, "(")
		key, _ := strconv.Atoi(element[0:i])
		if v[key] == "" {
			v[key] = element
		}
	}
	real := make([]string, 0)
	for index := range v {
		if v[index] != "" {
			real = append(real, v[index])
		}
	}
	return len(real)
}

//readFile read data
func readFile() *[][]string {
	iterators := make([][]string, 8, 8)
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
		iterators[i] = elements
	}
	timer <- time.Now().UnixNano()
	return &iterators
}

func getDataNormal(iterators *[][]string, channel *chan string) {
	for index := range *iterators {
		for i := range (*iterators)[index] {
			*channel <- (*iterators)[index][i]
		}
	}
	*channel <- ""
}
