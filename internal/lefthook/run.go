package lefthook

func Run(opts *Options, hookName string, gitArgs []string) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Run(hookName, gitArgs)
}

func (l *Lefthook) Run(hookName string, gitArgs []string) error {
	// Implement run command
	return nil
}
