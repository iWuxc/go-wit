package captcha

import "time"

type Fn func(opt *options)

type options struct {
	height   int
	width    int
	len      int
	maxSkew  float64
	dotCount int

	maxReties int

	expire time.Duration
}

// WithWidth sets the width of the captcha
func WithWidth(width int) Fn {
	return func(opt *options) {
		opt.width = width
	}
}

// WithHeight sets the height of the captcha
func WithHeight(height int) Fn {
	return func(opt *options) {
		opt.height = height
	}
}

// WithLength sets the length of captcha
func WithLength(len int) Fn {
	return func(opt *options) {
		opt.len = len
	}
}

// WithMaxSkew sets the max skew factor.
func WithMaxSkew(maxSkew float64) Fn {
	return func(opt *options) {
		opt.maxSkew = maxSkew
	}
}

// WithDotCount sets the number of dots to draw on the captcha
func WithDotCount(dotCount int) Fn {
	return func(opt *options) {
		opt.dotCount = dotCount
	}
}

// WithMaxRetires sets the max retries for captcha
// if the captcha is not solved in the given retries
// the captcha will be null
func WithMaxRetires(retires int) Fn {
	return func(opt *options) {
		if retires > 0 {
			opt.maxReties = retires
		}
	}
}

// WithExpire sets the life expire for captcha. default 10min
func WithExpire(duration time.Duration) Fn {
	return func(opt *options) {
		opt.expire = duration
	}
}
