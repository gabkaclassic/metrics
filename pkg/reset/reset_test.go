package reset

import "testing"

type testValue struct {
	n int
}

func (t *testValue) Reset() {
	t.n = 0
}

func TestResetablePool(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(p *ResetablePool[*testValue])
		action      func(p *ResetablePool[*testValue]) (*testValue, bool)
		expectNil   bool
		expectOK    bool
		expectValue *testValue
	}{
		{
			name: "get from empty pool",
			action: func(p *ResetablePool[*testValue]) (*testValue, bool) {
				return p.Get()
			},
			expectNil: true,
			expectOK:  false,
		},
		{
			name: "put resets value",
			setup: func(p *ResetablePool[*testValue]) {
				p.Put(&testValue{n: 42})
			},
			action: func(p *ResetablePool[*testValue]) (*testValue, bool) {
				return p.Get()
			},
			expectOK: true,
			expectValue: &testValue{
				n: 0,
			},
		},
		{
			name: "fifo order",
			setup: func(p *ResetablePool[*testValue]) {
				p.Put(&testValue{n: 1})
				p.Put(&testValue{n: 2})
			},
			action: func(p *ResetablePool[*testValue]) (*testValue, bool) {
				v, _ := p.Get()
				return v, true
			},
			expectOK: true,
			expectValue: &testValue{
				n: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New[*testValue]()

			if tt.setup != nil {
				tt.setup(p)
			}

			v, ok := tt.action(p)

			if ok != tt.expectOK {
				t.Fatalf("expected ok=%v, got %v", tt.expectOK, ok)
			}

			if tt.expectNil {
				if v != nil {
					t.Fatalf("expected nil, got %+v", v)
				}
				return
			}

			if tt.expectValue != nil {
				if v == nil {
					t.Fatalf("expected value, got nil")
				}
				if v.n != tt.expectValue.n {
					t.Fatalf("expected n=%d, got %d", tt.expectValue.n, v.n)
				}
			}
		})
	}
}

func TestResetablePool_Concurrent(t *testing.T) {
	tests := []struct {
		name       string
		workers    int
		iterations int
	}{
		{
			name:       "concurrent put/get",
			workers:    10,
			iterations: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New[*testValue]()
			done := make(chan struct{})

			for i := 0; i < tt.workers; i++ {
				go func() {
					for j := 0; j < tt.iterations; j++ {
						p.Put(&testValue{n: 100})
						_, _ = p.Get()
					}
					done <- struct{}{}
				}()
			}

			for i := 0; i < tt.workers; i++ {
				<-done
			}
		})
	}
}
