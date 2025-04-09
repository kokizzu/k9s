// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestStackClear(t *testing.T) {
	comps := []model.Component{makeC("c1"), makeC("c2"), makeC("c3")}
	uu := map[string]struct {
		items []model.Component
	}{
		"empty": {
			items: []model.Component{},
		},
		"items": {
			items: comps,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			s.Clear()
			assert.True(t, s.Empty())
		})
	}
}

func TestStackPrevious(t *testing.T) {
	comps := []model.Component{makeC("c1"), makeC("c2"), makeC("c3")}
	uu := map[string]struct {
		items []model.Component
		pops  int
		e     model.Component
	}{
		"empty": {
			items: []model.Component{},
			pops:  0,
			e:     nil,
		},
		"one_left": {
			items: comps,
			pops:  1,
			e:     comps[0],
		},
		"none_left": {
			items: comps,
			pops:  2,
			e:     comps[0],
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			for range u.pops {
				s.Pop()
			}
			assert.Equal(t, u.e, s.Previous())
		})
	}
}

func TestStackIsLast(t *testing.T) {
	uu := map[string]struct {
		items []model.Component
		pops  int
		e     bool
	}{
		"empty": {
			items: []model.Component{},
		},
		"normal": {
			items: []model.Component{makeC("c1"), makeC("c2"), makeC("c3")},
			pops:  1,
		},
		"last": {
			items: []model.Component{makeC("c1"), makeC("c2"), makeC("c3")},
			pops:  2,
			e:     true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			for range u.pops {
				s.Pop()
			}
			assert.Equal(t, u.e, s.IsLast())
		})
	}
}

func TestStackFlatten(t *testing.T) {
	uu := map[string]struct {
		items []model.Component
		e     []string
	}{
		"empty": {
			items: []model.Component{},
			e:     []string{},
		},
		"normal": {
			items: []model.Component{makeC("c1"), makeC("c2"), makeC("c3")},
			e:     []string{"c1", "c2", "c3"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			assert.Equal(t, u.e, s.Flatten())
			assert.Len(t, s.Peek(), len(u.e))
		})
	}
}

func TestStackPush(t *testing.T) {
	top := c{}
	uu := map[string]struct {
		items []model.Component
		pop   int
		e     bool
		top   model.Component
	}{
		"empty": {
			items: []model.Component{},
			pop:   3,
			e:     true,
		},
		"full": {
			items: []model.Component{c{}, c{}, top},
			pop:   3,
			e:     true,
		},
		"pop": {
			items: []model.Component{c{}, c{}, top},
			pop:   2,
			e:     false,
			top:   top,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, c := range u.items {
				s.Push(c)
			}
			for range u.pop {
				s.Pop()
			}
			assert.Equal(t, u.e, s.Empty())
			if !u.e {
				assert.Equal(t, u.top, s.Top())
			}
		})
	}
}

func TestStackTop(t *testing.T) {
	top := c{}
	uu := map[string]struct {
		items []model.Component
		e     model.Component
	}{
		"blank": {
			items: []model.Component{},
		},
		"push3": {
			items: []model.Component{c{}, c{}, top},
			e:     top,
		},
		"push1": {
			items: []model.Component{top},
			e:     top,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := model.NewStack()
			for _, item := range u.items {
				s.Push(item)
			}
			v := s.Top()
			assert.Equal(t, u.e, v)
		})
	}
}

func TestStackAddListener(t *testing.T) {
	items := []model.Component{c{}, c{}, c{}}
	s := model.NewStack()
	l := stackL{}
	s.AddListener(&l)
	for _, item := range items {
		s.Push(item)
	}
	assert.Equal(t, 3, l.count)

	for range items {
		s.Pop()
	}
	assert.Equal(t, 0, l.count)
}

func TestStackAddListenerAfter(t *testing.T) {
	items := []model.Component{c{}, c{}, c{}}
	s := model.NewStack()
	l := stackL{}
	for _, item := range items {
		s.Push(item)
	}
	s.AddListener(&l)
	assert.Equal(t, 1, l.tops)
	assert.Equal(t, 0, l.count)
}

func TestStackRemoveListener(t *testing.T) {
	s := model.NewStack()
	l1, l2, l3 := &stackL{}, &stackL{}, &stackL{}
	s.AddListener(l1)
	s.AddListener(l2)

	s.RemoveListener(l2)
	s.RemoveListener(l3)
	s.RemoveListener(l1)

	s.Push(c{})

	assert.Equal(t, 0, l1.count)
	assert.Equal(t, 0, l2.count)
	assert.Equal(t, 0, l3.count)
}

// ----------------------------------------------------------------------------
// Helpers...

type stackL struct {
	count, tops int
}

func (s *stackL) StackPushed(model.Component) {
	s.count++
}

func (s *stackL) StackPopped(_, _ model.Component) {
	s.count--
}

func (s *stackL) StackTop(model.Component) { s.tops++ }

type c struct {
	name string
}

func makeC(n string) c {
	return c{name: n}
}

func (c) InCmdMode() bool                                            { return false }
func (c c) Name() string                                             { return c.name }
func (c) SetCommand(*cmd.Interpreter)                                {}
func (c) Hints() model.MenuHints                                     { return nil }
func (c) HasFocus() bool                                             { return false }
func (c) ExtraHints() map[string]string                              { return nil }
func (c) Draw(tcell.Screen)                                          {}
func (c) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) { return nil }
func (c) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return nil
}
func (c) SetRect(int, int, int, int)       {}
func (c) GetRect() (a, b, c, d int)        { return 0, 0, 0, 0 }
func (c) GetFocusable() tview.Focusable    { return nil }
func (c) Focus(func(tview.Primitive))      {}
func (c) Blur()                            {}
func (c) Start()                           {}
func (c) Stop()                            {}
func (c) Init(context.Context) error       { return nil }
func (c) SetFilter(string)                 {}
func (c) SetLabelFilter(map[string]string) {}
