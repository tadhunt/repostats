all:
	go build repostats.go
	./repostats
	open nloc-over-time.svg
	open pcnt-chg-over-time.svg

fetch:
	go get code.google.com/p/plotinum/plot
	go get code.google.com/p/plotinum/plotter
	go get code.google.com/p/plotinum/plotutil
	go get code.google.com/p/plotinum/vg

data:
	-if [ -z "${REPO}" ] ; then echo "set REPO=/path/to/git/repo" 1>&2; exit 1; fi
	(cd ${REPO} && git log --reverse --numstat --date=short) | awk -f log-massage.awk | sort > commitdata.csv

clean:
	rm -f repostats *.svg *.png *.pdf commitdata.csv
