package termyx

import (
	"math"
	"sync"
	"time"
)

// ── Easing functions ─────────────────────────────────────────────────────────

// EasingFunc maps a normalized time t ∈ [0, 1] to a progress value ∈ [0, 1].
type EasingFunc func(t float64) float64

// Linear — constant velocity.
func Linear(t float64) float64 { return t }

// EaseIn — slow start, accelerates.
func EaseIn(t float64) float64 { return t * t }

// EaseOut — fast start, decelerates.
func EaseOut(t float64) float64 { return t * (2 - t) }

// EaseInOut — slow start and end, fast middle.
func EaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// EaseInCubic — cubic acceleration.
func EaseInCubic(t float64) float64 { return t * t * t }

// EaseOutCubic — cubic deceleration.
func EaseOutCubic(t float64) float64 {
	t--
	return t*t*t + 1
}

// EaseInOutCubic — cubic in-out.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return 0.5*t*t*t + 1
}

// Bounce simulates a bouncing ball easing out.
func Bounce(t float64) float64 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

// Elastic simulates a spring-overshoot easing out.
func Elastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	return math.Pow(2, -10*t)*math.Sin((t-p/4)*(2*math.Pi)/p) + 1
}

// ── Tween ─────────────────────────────────────────────────────────────────────

// Tween smoothly interpolates a float64 value over a fixed duration.
// Typical use: animate a progress bar, panel size, or numeric counter.
//
//	tw := termyx.NewTween(0, 100, 500*time.Millisecond, termyx.EaseOut)
//	tw.StartAutoTick(16*time.Millisecond, updateCh, stopCh)
//
//	// In Root:
//	termyx.ProgressBar(tw.Value(), " CPU", filled, empty)
type Tween struct {
	mu       sync.Mutex
	from, to float64
	ease     EasingFunc
	duration time.Duration
	start    time.Time
	running  bool
}

// NewTween creates a Tween from `from` to `to` over `duration` using `ease`.
func NewTween(from, to float64, duration time.Duration, ease EasingFunc) *Tween {
	if ease == nil {
		ease = EaseOut
	}
	return &Tween{from: from, to: to, duration: duration, ease: ease}
}

// Start begins the animation from the current time.
func (tw *Tween) Start() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.start = time.Now()
	tw.running = true
}

// SetTarget changes the destination and restarts the animation from the
// current interpolated value. Smooth when called mid-animation.
func (tw *Tween) SetTarget(to float64) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.from = tw.valueLocked()
	tw.to = to
	tw.start = time.Now()
	tw.running = true
}

// Value returns the current interpolated value. Returns `to` when done.
func (tw *Tween) Value() float64 {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	return tw.valueLocked()
}

func (tw *Tween) valueLocked() float64 {
	if !tw.running || tw.duration <= 0 {
		return tw.to
	}
	elapsed := time.Since(tw.start)
	if elapsed >= tw.duration {
		tw.running = false
		return tw.to
	}
	t := float64(elapsed) / float64(tw.duration)
	return tw.from + (tw.to-tw.from)*tw.ease(t)
}

// Done reports whether the animation has completed.
func (tw *Tween) Done() bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	return !tw.running
}

// StartAutoTick ticks the animation at the given interval, sending on notify
// to trigger re-renders. Close stopC to stop.
func (tw *Tween) StartAutoTick(interval time.Duration, notify chan<- struct{}, stopC <-chan struct{}) {
	tw.Start()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case notify <- struct{}{}:
				default:
				}
				if tw.Done() {
					return
				}
			case <-stopC:
				return
			}
		}
	}()
}

// ── AnimatedFloat ─────────────────────────────────────────────────────────────

// AnimatedFloat wraps a float64 that smoothly tracks a target value.
// When Set is called, it starts a Tween from the current value to the new target.
//
//	var cpu AnimatedFloat
//	cpu.Init(500*time.Millisecond, termyx.EaseOut)
//	cpu.Set(newValue, updateCh, stopCh)
type AnimatedFloat struct {
	mu       sync.Mutex
	tween    *Tween
	duration time.Duration
	ease     EasingFunc
}

// Init configures the animation duration and easing. Must be called before Set.
func (a *AnimatedFloat) Init(duration time.Duration, ease EasingFunc) {
	a.duration = duration
	a.ease = ease
}

// Set smoothly transitions to target, triggering re-renders via notify.
// Close stopC to cancel the goroutine.
func (a *AnimatedFloat) Set(target float64, notify chan<- struct{}, stopC <-chan struct{}) {
	a.mu.Lock()
	if a.tween == nil {
		a.tween = NewTween(0, target, a.duration, a.ease)
		a.mu.Unlock()
		a.tween.StartAutoTick(16*time.Millisecond, notify, stopC)
		return
	}
	a.tween.SetTarget(target)
	a.mu.Unlock()
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case notify <- struct{}{}:
				default:
				}
				if a.tween.Done() {
					return
				}
			case <-stopC:
				return
			}
		}
	}()
}

// Value returns the current animated value.
func (a *AnimatedFloat) Value() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tween == nil {
		return 0
	}
	return a.tween.Value()
}

// ── Counter ───────────────────────────────────────────────────────────────────

// Counter animates an integer count from its current value toward a target,
// incrementing or decrementing by step each tick. Useful for animated numbers
// in dashboards (request count, error count, etc.).
//
//	var req Counter
//	req.Set(newCount, updateCh, stopCh)
type Counter struct {
	mu       sync.Mutex
	current  int64
	target   int64
	step     int64
}

// Init sets the step size (default 1). Must be called before Set.
func (c *Counter) Init(step int64) {
	if step <= 0 {
		step = 1
	}
	c.step = step
}

// Set transitions the counter toward target, sending on notify each tick.
func (c *Counter) Set(target int64, interval time.Duration, notify chan<- struct{}, stopC <-chan struct{}) {
	c.mu.Lock()
	c.target = target
	step := c.step
	if step == 0 {
		step = 1
	}
	c.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				if c.current < c.target {
					c.current += step
					if c.current > c.target {
						c.current = c.target
					}
				} else if c.current > c.target {
					c.current -= step
					if c.current < c.target {
						c.current = c.target
					}
				}
				done := c.current == c.target
				c.mu.Unlock()
				select {
				case notify <- struct{}{}:
				default:
				}
				if done {
					return
				}
			case <-stopC:
				return
			}
		}
	}()
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

// SetImmediate jumps the counter to value without animation.
func (c *Counter) SetImmediate(value int64) {
	c.mu.Lock()
	c.current = value
	c.target = value
	c.mu.Unlock()
}
