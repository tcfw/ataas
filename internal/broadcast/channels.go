package broadcast

import "sync"

//Merge merges multiple chans together
func Merge(cs ...<-chan []byte) <-chan []byte {
	out := make(chan []byte)

	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan []byte) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
