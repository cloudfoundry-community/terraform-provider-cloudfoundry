package common

import (
	"fmt"
	"time"
)

func Polling(pollingFunc func() (bool, error), waitTime time.Duration) error {

	for {
		finished, err := pollingFunc()
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		time.Sleep(waitTime)
	}
	return nil
}

func PollingWithTimeout(pollingFunc func() (bool, error), waitTime time.Duration, timeout time.Duration) error {
	stagingStartTime := time.Now()
	for {
		if time.Since(stagingStartTime) > timeout {
			return fmt.Errorf("Timeout reached")
		}
		finished, err := pollingFunc()
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		time.Sleep(waitTime)
	}
	return nil
}
