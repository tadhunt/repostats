all:
	go build repostats.go

fetch:
	go get code.google.com/p/plotinum/plot
	go get code.google.com/p/plotinum/plotter
	go get code.google.com/p/plotinum/plotutil
	go get code.google.com/p/plotinum/vg

#
# LC_ALL=C	Fixes processing of the Linux commit history
#
data:
	@if [ -z "${REPO}" ] ; then echo "set REPO=/path/to/git/repo" 1>&2; exit 1; fi
	(cd ${REPO} && git log --reverse --numstat --date=short) | awk -f log-massage.awk | LC_ALL=C sort > commitdata.csv

clean:
	rm -f repostats *.svg *.png *.pdf commitdata.csv

#
# Valid linux data seems to start about 1,309 days from the first commit
# Note that 
linux: all linux/linux-commitdata.csv
	./repostats \
		-infile=linux/linux-commitdata.csv \
		-nloc="linux/nloc-over-time.pdf" \
		-pcnt="linux/pcnt-chg-over-time.pdf" \
		-pspan=7 \
		-codebase="the Linux kernel" \
		-firstday=1309
	open linux/nloc-over-time.pdf linux/pcnt-chg-over-time.pdf

#
# Linux kernel commit history has some commits with screwed up dates:
#
#	1970-01-01	224426f168aa4af3dcb628e6edaa824d32d60e6f 
#	2019-04-05	a27ac38efd6dc6dccebfc9bcc475ab4aa5fc4a56
#	2030-08-14	09f2724a786f76475ef2985cf84f5359c553aade
#	2037-04-25	12ca45fea91cfbb09df828bea958b47348caee6d
#
# So this funkiness elides these commits from the data

linux/linux-commitdata.csv:
	test -d linux || mkdir linux
	${MAKE} data
	awk '
		/224426f168aa4af3dcb628e6edaa824d32d60e6f/ ||
		/a27ac38efd6dc6dccebfc9bcc475ab4aa5fc4a56/ ||
		/09f2724a786f76475ef2985cf84f5359c553aade/ ||
		/12ca45fea91cfbb09df828bea958b47348caee6d/ {
			next
		}
		{
			 print $0
		}
	' commitdata.csv > linux/linux-commitdata.csv
