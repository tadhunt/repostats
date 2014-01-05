#!/bin/awk
#
# This script parses the git log output from the following git command, and turns it into csv
#
#	git log --reverse --numstat --date=short
#
BEGIN {
	stderr = "/dev/stderr"

#	printf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
#		"Date",
#		"Number of Insertions",
#		"Number of Deletions",
#		"Author",
#		"Filename")

	last = "none"
}

/^commit[ 	].*/ {
	commit = $2

	last = "commit"
	next
}

/^Author:[ 	].*/ {
	author = $0
	sub(/^Author: /, "", author)

	last = "author"
	next
}

/^Date:[ 	].*/ {
	date = $2

	last = "date"
	next
}

/^Merge:[ 	].*/ {
	ismerge = 1
	last = "merge"
	next
}

/^[ 	]+.*/ {
	last = "logmsg"
	next
}

/^-[ 	]+-[ 	]+.*/ {
	last = "binary"
	next
}

/^[0-9]+[ 	]+[0-9]+[ 	]+.*/ {
	ninsertions = $1
	ndeletions = $2
	file = $0
	sub(/^[0-9]+[ 	]+[0-9]+[ 	]+/, "", file)	# easier this way because the filename might contain spaces

	gsub(/"/, "'", file)
	gsub(/"/, "'", author)

	if(ninsertions > 0 || ndeletions > 0) {
		printf("\"%s\",%d,%d,\"%s\",\"%s\",\"%s\"\n",
			date,
			ninsertions,
			ndeletions,
			commit,
			author,
			file)
	}

	last = "nlines"
	next
}

/^$/ {
	if(last == "date")
		next
	if(last == "logmsg")
		next
	if(last == "nlines" || last == "binary") {
		commit = ""
		author = ""
		date=""
		skip = 0
		ninsertions = 0
		ndeletions = 0
		file = ""
		ismerge = 0
		last = "blank"
		next
	}

	printf("ERROR: unexpected blank line %d last %s\n", NR, last) > stderr
	exit 1
}

{
	printf("ERROR: Line %d: unhandled\n", NR) > stderr
	printf("ERROR: \"%s\"\n", $0) > stderr
	exit 1
}
