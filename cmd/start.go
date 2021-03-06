package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/coreos/coreinit/job"
	"github.com/coreos/coreinit/unit"
)

func newStartUnitCommand() cli.Command {
	return cli.Command{
		Name: "start",
		Flags: []cli.Flag{
			cli.StringFlag{"require", "", "Filter hosts with a set of requirements. Format is comma-delimited list of <key>=<value> pairs."},
		},
		Usage:  "Start (activate) one or more units",
		Action: startUnitAction,
	}
}

func startUnitAction(c *cli.Context) {
	r := getRegistry(c)

	cliRequirements := parseRequirements(c.String("require"))

	payloads := make([]job.JobPayload, len(c.Args()))
	for i, v := range c.Args() {
		out, err := ioutil.ReadFile(v)
		if err != nil {
			fmt.Printf("%s: No such file or directory\n", v)
			return
		}

		unitFile := unit.NewSystemdUnitFile(string(out))
		fileRequirements := unitFile.Requirements()
		requirements := stackRequirements(fileRequirements, cliRequirements)

		name := path.Base(v)
		payload, err := job.NewJobPayload(name, unitFile.String(), requirements)
		if err != nil {
			fmt.Println(err.Error())
			return
		} else {
			payloads[i] = *payload
		}
	}

	req, err := job.NewJobRequest(payloads)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	r.AddRequest(req)
}

func parseRequirements(arg string) map[string][]string {
	reqs := make(map[string][]string, 0)

	add := func(key, val string) {
		vals, ok := reqs[key]
		if !ok {
			vals = make([]string, 0)
			reqs[key] = vals
		}
		vals = append(vals, val)
		reqs[key] = vals
	}

	for _, pair := range strings.Split(arg, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		add(key, val)
	}

	return reqs
}

func stackRequirements(base, overlay map[string][]string) map[string][]string{
	stacked := make(map[string][]string, 0)

	for key, values := range base {
		stacked[key] = values
	}

	for key, values := range overlay {
		stacked[key] = values
	}

	return stacked
}
