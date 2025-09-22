package nlreturnfmt

type Option func(*Formatter)

func WithBlockSize(blockSize int) Option {
	return func(f *Formatter) {
		if blockSize >= 0 {
			f.blockSize = blockSize
		}
	}
}

func WithWrite() Option {
	return func(f *Formatter) { f.write = true }
}

func WithDryRun() Option {
	return func(f *Formatter) { f.dryRun = true }
}

func WithVerbose() Option {
	return func(f *Formatter) { f.verbose = true }
}

func WithParallelism(n int) Option {
	return func(f *Formatter) {
		if n > 0 {
			f.parallelism = n
		}
	}
}
