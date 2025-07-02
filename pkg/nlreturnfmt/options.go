package nlreturnfmt

type Option func(*Formater)

func WithBlockSize(blockSize uint) Option {
	return func(f *Formater) { f.blockSize = blockSize }
}

func WithWrite() Option {
	return func(f *Formater) { f.write = true }
}

func WithDryRun() Option {
	return func(f *Formater) { f.dryRun = true }
}

func WithVerbose() Option {
	return func(f *Formater) { f.verbose = true }
}
