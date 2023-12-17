package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)


type Response struct {
	SortedArrays [][]int `json:"sorted arrays"`
	TimeNS       string   `json:"time ns"`
}


func sequentialSort(arr []int) []int {
	sortedArr := make([]int, len(arr))
	copy(sortedArr, arr)
	sort.Ints(sortedArr)
	return sortedArr
}


func concurrentSort(arr []int, wg *sync.WaitGroup, ch chan []int) {
	defer wg.Done()
	sortedArr := make([]int, len(arr))
	copy(sortedArr, arr)
	sort.Ints(sortedArr)
	ch <- sortedArr
}


func process(w http.ResponseWriter, r *http.Request, sortingFunc func([]int) []int) {
	var input struct {
		ToSort [][]int `json:"to_sort"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	var sortedArrays [][]int
	var wg sync.WaitGroup
	ch := make(chan []int)

	for _, subArray := range input.ToSort {
		wg.Add(1)
		go func(arr []int) {
			defer wg.Done()
			sorted := sortingFunc(arr)
			ch <- sorted
		}(subArray)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for sortedSubArray := range ch {
		sortedArrays = append(sortedArrays, sortedSubArray)
	}

	duration := time.Since(startTime)

	responseData := Response{
		SortedArrays: sortedArrays,
		TimeNS:       fmt.Sprint(duration.Nanoseconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func singleProcess(w http.ResponseWriter, r *http.Request) {
	process(w, r, sequentialSort)
}

func concurrentProcess(w http.ResponseWriter, r *http.Request) {
	process(w, r, func(arr []int) []int {
		var wg sync.WaitGroup
		ch := make(chan []int)

		wg.Add(1)
		go concurrentSort(arr, &wg, ch)

		go func() {
			wg.Wait()
			close(ch)
		}()

		return <-ch
	})
}

func main() {
	http.HandleFunc("/process-single", singleProcess)
	http.HandleFunc("/process-concurrent", concurrentProcess)

	fmt.Println("Server is listening on port 8000")
	http.ListenAndServe(":8000", nil)
}
