package main

import (
	"github.com/pkg/errors"
	"github.com/projectatomic/libpod/libpod"
	"github.com/urfave/cli"
)

var (
	attachFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "detach-keys",
			Usage: "Override the key sequence for detaching a container. Format is a single character [a-Z] or ctrl-<value> where <value> is one of: a-z, @, ^, [, , or _.",
		},
		cli.BoolFlag{
			Name:  "no-stdin",
			Usage: "Do not attach STDIN. The default is false.",
		},
		LatestFlag,
	}
	attachDescription = "The podman attach command allows you to attach to a running container using the container's ID or name, either to view its ongoing output or to control it interactively."
	attachCommand     = cli.Command{
		Name:        "attach",
		Usage:       "Attach to a running container",
		Description: attachDescription,
		Flags:       attachFlags,
		Action:      attachCmd,
		ArgsUsage:   "",
	}
)

func attachCmd(c *cli.Context) error {
	args := c.Args()
	var ctr *libpod.Container
	if err := validateFlags(c, attachFlags); err != nil {
		return err
	}
	if len(c.Args()) > 1 || (len(c.Args()) == 0 && !c.Bool("latest")) {
		return errors.Errorf("attach requires the name or id of one running container or the latest flag")
	}

	runtime, err := getRuntime(c)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.Shutdown(false)

	if c.Bool("latest") {
		ctr, err = runtime.GetLatestContainer()
	} else {
		ctr, err = runtime.LookupContainer(args[0])
	}

	if err != nil {
		return errors.Wrapf(err, "unable to exec into %s", args[0])
	}

	conState, err := ctr.State()
	if err != nil {
		return errors.Wrapf(err, "unable to determine state of %s", args[0])
	}
	if conState != libpod.ContainerStateRunning {
		return errors.Errorf("you can only attach to running containers")
	}

	if err := ctr.Attach(c.Bool("no-stdin"), c.String("detach-keys")); err != nil {
		return errors.Wrapf(err, "error attaching to container %s", ctr.ID())
	}

	return nil
}
