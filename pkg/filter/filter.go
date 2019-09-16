package filter

type Filter interface {
	Name() string
	Init() error
	Pre(Context) (statusCode int, err error)
	Route(Context) (statusCode int, err error)
	Post(Context) (statusCode int, err error)
}

type Context interface {
}
