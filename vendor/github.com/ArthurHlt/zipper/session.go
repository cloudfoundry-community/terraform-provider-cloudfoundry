package zipper

type Session struct {
	handler Handler
	src     *Source
}

// Create a new session
func NewSession(src *Source, handler Handler) *Session {
	return &Session{handler, src}
}

// Create zip file
func (s Session) Zip() (ZipReadCloser, error) {
	return s.handler.Zip(s.src)
}

// Retrieve signature
func (s Session) Sha1() (string, error) {
	return s.handler.Sha1(s.src)
}

// Check if source signature is different from a previous signature
// If true, it's mean than files have changed
func (s Session) IsDiff(storedSha1 string) (bool, string, error) {
	sha1Given, err := s.Sha1()
	if err != nil {
		return true, "", err
	}
	return storedSha1 != sha1Given, sha1Given, nil
}

// Retrieve handler use in the session
func (s Session) Handler() Handler {
	return s.handler
}

// Retrieve source use in the session
func (s Session) Source() *Source {
	return s.src
}
