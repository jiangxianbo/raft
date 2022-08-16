package xuanxiangmoshi

// 选项模式

import "fmt"

type Options struct {
	strOption1 string
	strOption2 string
	strOption3 string
	intOption1 int
	intOption2 int
	intOption3 int
}

type Option func(opts *Options)

func InitOptions1(opts ...Option) {
	option := &Options{}
	for _, opt := range opts {
		// 调用函数，在函数内部，给传递的对象赋值
		opt(option)
	}
	fmt.Printf("init options %#v\n", option)
}

func WithStrOption1(str string) Option {
	return func(opts *Options) {
		opts.strOption1 = str
	}
}

func WithStrOption2(str string) Option {
	return func(opts *Options) {
		opts.strOption2 = str
	}
}

func WithStrOption3(str string) Option {
	return func(opts *Options) {
		opts.strOption3 = str
	}
}

func WithIntOption1(i int) Option {
	return func(opts *Options) {
		opts.intOption1 = i
	}
}

func WithIntOption2(i int) Option {
	return func(opts *Options) {
		opts.intOption2 = i
	}
}

func WithIntOption3(i int) Option {
	return func(opts *Options) {
		opts.intOption3 = i
	}
}

func xxx() {
	InitOptions1(WithStrOption1("str1"), WithIntOption2(2))
}
