BINDIR=${HOME}/bin

all:
	go build -v repostats.go

install: all
	rm -f ${BINDIR}/repostats && cp repostats ${BINDIR}

clean:
	go clean -cache -modcache
	rm -f repostats *.svg *.png *.pdf commitdata.csv

uninstall:
	rm -f ${BINDIR}/repostats

fetch:
	go mod init github.com/tadhunt/repostats
	go mod tidy

#
# LC_ALL=C	Fixes processing of the Linux commit history
#
data: data-fetch data-show

data-fetch:
	@if [ -z "${REPO}" ] ; then echo "set REPO=/path/to/git/repo" 1>&2; exit 1; fi
	(cd ${REPO} && git log --reverse --numstat --date=short) | awk -f log-massage.awk | LC_ALL=C sort > commitdata.csv

data-show:
	${BINDIR}/repostats \
		-infile=commitdata.csv \
		-nloc="nloc-over-time.pdf" \
		-pcnt="pcnt-chg-over-time.pdf" \
		-pspan=1 \
		-codebase="${REPO}"
		-firstday=0
	open nloc-over-time.pdf pcnt-chg-over-time.pdf



selfdata:
	git log --reverse --numstat --date=short | awk -f log-massage.awk | LC_ALL=C sort > commitdata.csv
	${BINDIR}/repostats \
		-infile=commitdata.csv \
		-nloc="nloc-over-time.pdf" \
		-pcnt="pcnt-chg-over-time.pdf" \
		-pspan=1 \
		-codebase="The Repostats Repository"
		-firstday=0
	open nloc-over-time.pdf pcnt-chg-over-time.pdf

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

commitdata.csv:
	${MAKE} data

linux/linux-commitdata.csv: commitdata.csv
	test -d linux || mkdir linux
	awk '	/224426f168aa4af3dcb628e6edaa824d32d60e6f/ ||	\
		/a27ac38efd6dc6dccebfc9bcc475ab4aa5fc4a56/ ||	\
		/09f2724a786f76475ef2985cf84f5359c553aade/ ||	\
		/12ca45fea91cfbb09df828bea958b47348caee6d/ {	\
			next					\
		}						\
		{						\
			 print $0				\
		}						\
	' commitdata.csv > linux/linux-commitdata.csv
