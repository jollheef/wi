# Tiny¹ non-interactive cli browser

[1] Actually, that's not true anymore.

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
