package limiter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type mockHandler struct {
	mock.Mock
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called(w, r)
	time.Sleep(2 * time.Second)
}

func TestNew(t *testing.T) {
	maxReqs := 100
	dropOverLimit := true
	limiter := New(maxReqs, &mockHandler{}, dropOverLimit)
	assert.Equal(t, maxReqs, cap(limiter.reqs))
	assert.Equal(t, dropOverLimit, limiter.dropOverLimit)
}

func TestLimiter_ServeHTTP(t *testing.T) {
	tests := []struct {
		name          string
		maxReqs       int
		dropOverLimit bool
	}{
		{
			name:          "drop over limit",
			maxReqs:       2,
			dropOverLimit: true,
		},
		{
			name:          "do not drop over limit",
			maxReqs:       2,
			dropOverLimit: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &mockHandler{}
			limiter := New(tt.maxReqs, h, tt.dropOverLimit)
			// we expect a non-blocking ServerHTTP to be called tt.maxReqs number of times
			nonBlockedCalls := h.On("ServeHTTP", mock.Anything, mock.Anything).Times(tt.maxReqs)
			var lastCallTime, lastReturnTime time.Time
			if !tt.dropOverLimit {
				// if limiter is configured to not drop limit-exceeding requests then we expect another call to
				// ServeHTTP but not before the previous ones. This one will be delayed due to the limit.
				h.On("ServeHTTP", mock.Anything, mock.Anything).NotBefore(nonBlockedCalls).Run(func(_ mock.Arguments) {
					lastReturnTime = time.Now() // record the time of the delayed request actual execution
				})
			}
			var wg sync.WaitGroup
			wg.Add(tt.maxReqs + 1)
			for i := 0; i < tt.maxReqs+1; i++ {
				lastCallTime = time.Now() // this will eventually record the time of the delayed request start
				go func() {
					defer wg.Done()
					limiter.ServeHTTP(httptest.NewRecorder(), nil)
				}()
			}
			wg.Wait()
			if !tt.dropOverLimit {
				// if limiter is configured to not drop limit-exceeding requests then we expect the time delta between
				// the request start and its execution to be at least the amount of time the mock request handler sleeps
				// for [line:19]
				delta := lastReturnTime.Sub(lastCallTime)
				t.Log(delta)
				assert.GreaterOrEqual(t, delta, 2*time.Second)
			}
			h.AssertExpectations(t)
		})
	}
}
