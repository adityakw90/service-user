package service

import (
	"context"
	"sync"
)

func fetchDataWithContext(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan error,
	fetchFunc func() (interface{}, error),
	processFunc func(interface{}),
) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return
	default:
		// Execute the fetch function
		result, err := fetchFunc()
		if err != nil {
			errChan <- err
			return
		}

		// Process the fetched result
		processFunc(result)

		// Signal no error occurred
		errChan <- nil
	}
}
