package options

type Waiter interface {
	Wait() error
	Add(resources ...string) Waiter
}
