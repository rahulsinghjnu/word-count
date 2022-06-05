package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/rahulsinghjnu/word-count/service"
	"github.com/rahulsinghjnu/word-count/util"
)

/**
	Author: @Rahul Singh
**/

// Define Flag
var help = flag.Bool("help", false, "Show help")
var filePath = ""
var topN = 10

type wordCountService struct {
	r            *bufio.Reader
	respChan     *chan map[string]int
	doneChannel  *chan struct{}
	semaphore    *chan struct{}
	linesPool    *sync.Pool
	wg           *sync.WaitGroup
	wcServiceURL string
}

func main() {
	// Bind the flag
	flag.StringVar(&filePath, "filePath", "", "Complete File Path")
	flag.IntVar(&topN, "topN", 10, "Top N frequency words")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if filePath == "" {
		fmt.Print("Please mention the filePath as command line argument and use command go run main.go -filePath=<FilePath>. Use -help to get Help. ")
		os.Exit(0)
	}

	err := godotenv.Load("local.env")
	if err != nil {
		panic("Env file loading error.")
	}

	url := os.Getenv("WORD_COUNT_SERVICE_URL")
	if url == "" {
		fmt.Println("WORD_COUNT_SERVICE_URL in local.env is empty or not mentioned")
		os.Exit(0)
	}
	// Open file to read the string
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}
	// Once text processing is done, close the channel.
	defer file.Close()

	// Buffer IO reader to read the file.
	r := bufio.NewReader(file)

	// Pool for []byte of 64 KB.
	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, 64*1024)
		return lines
	}}

	var wg sync.WaitGroup

	// word count service response to be sent for processing
	respChan := make(chan map[string]int)

	// Done channel to signal the completion of response merging.
	doneChan := make(chan struct{})

	// maxParallel denotes the number of concurrent rest api execution.
	maxParallel := 5
	semaphore := make(chan struct{}, maxParallel)

	wcs := wordCountService{
		r:            r,
		respChan:     &respChan,
		doneChannel:  &doneChan,
		semaphore:    &semaphore,
		linesPool:    &linesPool,
		wg:           &wg,
		wcServiceURL: url,
	}

	wcs.readFileAndProcess()

	wc := make(map[string]int)
	lock := &sync.Mutex{}

	// main go routine to wait for completion for others go routine via doneChan
	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	isCompleted := false
	// Merge the service response to get the word frequency in whole file.
	for !isCompleted {
		select {
		case l := <-respChan:
			// critical section
			// Use the lock to increment the word count
			lock.Lock()
			for key, val := range l {
				wc[key] += val
			}
			lock.Unlock()
			wg.Done()
		case <-doneChan:
			isCompleted = true
		}
	}

	// Close the channels
	wcs.close()
	// Get top N words
	tonNWords := getTopNWords(wc)
	fmt.Println(tonNWords)
}

// Read large size file into 64KB chunks of data and sent it for processing
func (wcs *wordCountService) readFileAndProcess() error {
	for {
		// Get the byte array from pool
		buf := wcs.linesPool.Get().([]byte)

		// Read up to 64 KB of data
		n, err := wcs.r.Read(buf[:cap(buf)])
		buf = buf[:n]

		// Read Until EOF
		if n == 0 {
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println(err)
				return err
			}

		}

		wcs.wg.Add(1)
		// Asynchronously process the chunk of text to ger word frequency
		go func() {
			wcs.processChunk(buf)
		}()
	}
	return nil
}

func (wcs *wordCountService) processChunk(chunk []byte) {
	*wcs.semaphore <- struct{}{}
	// Serivce call to get word frequency.
	// TODO: Error handling
	wcResp, _ := service.GetWordCount(wcs.wcServiceURL, chunk)
	// Write the resp into channel to be processed.
	*wcs.respChan <- wcResp
	// put back []byte into the pool
	wcs.linesPool.Put(chunk)
	// Decrease the uses of semaphore count so that other goroutine can call the service.
	<-*wcs.semaphore
}

func (wcs *wordCountService) close() {
	// Close the channels and semaphore
	close(*wcs.respChan)
	close(*wcs.doneChannel)
	close(*wcs.semaphore)
}

// Func to get configured number of top N words based on frequency
func getTopNWords(wc map[string]int) string {
	sortedWordFrequencies := util.RankByWordCount(wc)
	topNWords := sortedWordFrequencies[:topN]
	topNWordsJson, _ := json.Marshal(&topNWords)
	return string(topNWordsJson)
}
