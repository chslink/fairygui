package fgui

import (
	"context"
	"time"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

// LoadPackage loads a FairyGUI package from the given path with context cancellation.
func LoadPackage(ctx context.Context, loader Loader, path string) (*Package, error) {
	type result struct {
		pkg *Package
		err error
	}
	ch := make(chan result, 1)
	go func() {
		data, err := loader.LoadOne(ctx, path, ResourceBinary)
		if err != nil {
			ch <- result{err: err}
			return
		}
		pkg, err := ParsePackage(data, path)
		ch <- result{pkg: pkg, err: err}
	}()
	select {
	case r := <-ch:
		return r.pkg, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// AdvanceWithContext drives the root stage with cancellation support.
func AdvanceWithContext(ctx context.Context, delta time.Duration, mouse MouseState) {
	if ctx.Err() != nil {
		return
	}
	core.Inst().Advance(delta, mouse)
}

// WaitForTransition plays a transition and blocks until complete or context cancelled.
func WaitForTransition(ctx context.Context, t *Transition, times int, delay float64) error {
	if t == nil {
		return nil
	}
	done := make(chan struct{}, 1)
	t.Play(times, delay)
	go func() {
		// Poll until transition completes
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(50 * time.Millisecond)
				if !t.Playing() {
					done <- struct{}{}
					return
				}
			}
		}
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		t.Stop(true)
		return ctx.Err()
	}
}
