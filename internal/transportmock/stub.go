package transportmock

type stub struct {
	matched   bool
	matcher   Matcher
	name      string
	responder Responder
}

func (s stub) String() string {
	return s.name
}
