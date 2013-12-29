#!/bin/awk
# commit 989fb6102add218589e8779c76fd61f91e113684
# Author: Tad Hunt <tadhunt@tad-hunts-macbook-pro.local>
# Date:   2009-11-30
# 
#     ignore objects
# 
# 4       0       .gitignore
# 
# commit 54f97a5baf887f0d2a73745efaa0a647ed20512e
# Author: Tad Hunt <tadhunt@tad-hunts-macbook-pro.local>
# Date:   2009-11-30
# 
#     reorg & add pitch
# 
# 0       46      Makefile
# 0       62      bwcalc.c
# 0       32      dedupgain.sh
# 0       56      dedupgaintree.sh
# 0       391     sha1block.c
# 0       73      sha1speed.c
# 46      0       src/Makefile
# 62      0       src/bwcalc.c
# 32      0       src/dedupgain.sh
# 56      0       src/dedupgaintree.sh
# 391     0       src/sha1block.c
# 73      0       src/sha1speed.c
# 
# commit 509aff7e53df7406749274d943ed573a0a78af02
# ...
# 
BEGIN {
	stderr = "/dev/stderr"

#	printf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
#		"Date",
#		"Number of Insertions",
#		"Number of Deletions",
#		"Author",
#		"Date",
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
	next						# don't track merge history
}

/^[ 	]+.*/ {
	last = "logmsg"
	next						# don't track log message
}

/^-[ 	]+-[ 	]+.*/ {
	last = "binary"
	next						# don't track binary files
}

/^[0-9]+[ 	]+[0-9]+[ 	]+.*/ {
	ninsertions = $1
	ndeletions = $2
	file = $0
	sub(/^[0-9]+[ 	]+[0-9]+[ 	]+/, "", file)	# easier this way because the filename might contain spaces
	gsub(/"/, "'", file)

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
