package zipper

type Handler interface {
	Zip(src *Source) (zip ZipReadCloser, err error)
	Sha1(src *Source) (sha1 string, err error)
	Detect(src *Source) bool
	Name() string
}
