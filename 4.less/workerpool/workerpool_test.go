package workerpool

import (
	"errors"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []Task
		n             int
		m             int
		expectedError error
	}{
		{
			name: "All tasks successful",
			tasks: []Task{
				func() error { time.Sleep(10 * time.Millisecond); return nil },
				func() error { time.Sleep(20 * time.Millisecond); return nil },
				func() error { time.Sleep(30 * time.Millisecond); return nil },
			},
			n:             2,
			m:             1,
			expectedError: nil,
		},
		{
			name: "Error limit exceeded",
			tasks: []Task{
				func() error { return errors.New("error 1") },
				func() error { return errors.New("error 2") },
				func() error { return nil },
			},
			n:             2,
			m:             1,
			expectedError: ErrErrorsLimitExceeded,
		},
		{
			name: "M is zero",
			tasks: []Task{
				func() error { return nil },
			},
			n:             1,
			m:             0,
			expectedError: ErrErrorsLimitExceeded,
		},
		{
			name: "M is negative",
			tasks: []Task{
				func() error { return nil },
			},
			n:             1,
			m:             -1,
			expectedError: ErrErrorsLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := Run(tt.tasks, tt.n, tt.m)
				if err != tt.expectedError {
					t.Errorf("Run() error = %v, expectedError %v", err, tt.expectedError)
				}
			},
		)
	}
}
