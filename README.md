[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1c1b7414632b439eb6e1ffab0806c541)](https://app.codacy.com/app/jollheef/wi?utm_source=github.com&utm_medium=referral&utm_content=jollheef/wi&utm_campaign=Badge_Grade_Dashboard)
[![Build Status](https://travis-ci.org/jollheef/wi.svg?branch=master)](https://travis-ci.org/jollheef/wi)

# Wi: Non-interactive CLI browser with embedded Tor

## Installation

    go get -u code.dumpstack.io/tools/wi

## Usage

	usage: wi [<flags>] <command> [<args> ...]

	Flags:
	  --help  Show context-sensitive help (also try --help-long and --help-man).
	  --tor   Use embedded tor

	Commands:
	  help [<command>...]
		Show help.

	  get <url>
		Get url

	  form <id> [<args>...]
		Fill form

	  link [<flags>] <no>
		Get link

		--history  Item from history

	  history [<flags>] [<items>]
		List history

		--all  Show all items

	  search [<string>...]
		Search by duckduckgo
