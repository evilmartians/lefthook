package lefthook

func Install(opts *Options, args *InstallArgs) error {
	return New(opts).Install(args)
}

func Uninstall(opts *Options, args *UninstallArgs) error {
	return New(opts).Uninstall(args)
}
