package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

type keyVal struct {
	value string
	key   int64
}

var mutex = sync.Mutex{}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	jobsWg := sync.WaitGroup{}

	for _, j := range jobs {
		out := make(chan interface{})
		job := j
		jobsWg.Add(1)

		go func(in, out chan interface{}) {
			job(in, out)
			close(out)
			jobsWg.Done()
		}(in, out)

		in = out
	}

	jobsWg.Wait()
}

func SingleHash(in chan interface{}, out chan interface{}) {
	jobWg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		jobWg.Add(1)

		go func() {
			for data := range in {
				results := make([]string, 2)
				val, success := data.(string)

				if !success {
					val = strconv.FormatInt(int64(data.(int)), 10)
				}

				mutex.Lock()
				values := []string{
					val,
					DataSignerMd5(val),
				}
				mutex.Unlock()

				tempChan := make(chan keyVal)
				wg := sync.WaitGroup{}

				for k, value := range values {
					key := int64(k)
					wg.Add(1)
					go func(tempCh chan keyVal, val string, num int64) {
						tempCh <- keyVal{
							value: DataSignerCrc32(val),
							key:   num,
						}

						wg.Done()
					}(tempChan, value, key)
				}

				go func() {
					wg.Wait()
					close(tempChan)
				}()

				for i := range tempChan {
					results[i.key] = i.value
				}

				out <- strings.Join(results, "~")
			}
			jobWg.Done()
		}()
	}

	jobWg.Wait()

	return
}

func MultiHash(in chan interface{}, out chan interface{}) {
	jobWg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		jobWg.Add(1)

		go func() {
			for data := range in {
				thSlice := []string{"0", "1", "2", "3", "4", "5"}

				initVal := data.(string)
				results := make([]string, 6)
				tempChan := make(chan keyVal, 1)
				wg := sync.WaitGroup{}

				for _, th := range thSlice {
					wg.Add(1)

					go func(tempCh chan keyVal, th string) {
						val := DataSignerCrc32(th + initVal)
						num, err := strconv.ParseInt(th, 10, 64)

						if err != nil {
							panic(err)
						}

						data := keyVal{
							value: val,
							key:   num,
						}

						tempCh <- data
						wg.Done()
					}(tempChan, th)
				}

				go func() {
					wg.Wait()
					close(tempChan)
				}()

				for val := range tempChan {
					results[val.key] = val.value
				}

				out <- strings.Join(results, "")
			}

			jobWg.Done()
		}()
	}
	jobWg.Wait()

	return
}

func CombineResults(in chan interface{}, out chan interface{}) {
	result := make([]string, 0, 6)

	for input := range in {
		val := input.(string)
		result = append(result, val)
		sort.Strings(result)
	}

	out <- strings.Join(result, "_")
}
