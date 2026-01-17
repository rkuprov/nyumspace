package daemon

type DaemonOpt func(*Daemon) error

// WithAddress sets the address for the daemon server.
func WithAddress(addr string) DaemonOpt {
	return func(d *Daemon) error {
		d.Server.Addr = addr
		return nil
	}
}
