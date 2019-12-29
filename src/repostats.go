package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
//	"code.google.com/p/plotinum/plot"
//	"code.google.com/p/plotinum/plotter"
//	"code.google.com/p/plotinum/plotutil"
//	"code.google.com/p/plotinum/vg"
)

type Record struct {
	Date   string
	Nins   int64
	Ndel   int64
	Commit string
	Author string
	File   string
}

type Nloc struct {
	Date time.Time
	Nloc int64
}

func parsedate(s string) (time.Time, error) {
	var err error

	a := strings.Split(s, "-")

	if len(a) != 3 {
		return time.Now(), errors.New("time format syntax error, expected YYYY-MM-DD")
	}

	var year, month, day uint64
	year, err = strconv.ParseUint(a[0], 10, 32)
	if err != nil {
		return time.Now(), err
	}

	month, err = strconv.ParseUint(a[1], 10, 32)
	if err != nil {
		return time.Now(), err
	}

	day, err = strconv.ParseUint(a[2], 10, 32)
	if err != nil {
		return time.Now(), err
	}

	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Local), nil
}

func reader_setup(inpath string) (*csv.Reader, error) {
	file, err := os.Open(inpath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)

	return reader, nil
}

//
// Parses csv records with the following columns:
//
//	date, number-of-lines-inserted, number-of-lines-deleted, commit-hash, author, filename
//
func record_get(infile *csv.Reader) (*Record, error) {
	line, err := infile.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(line) < 6 {
		return nil, errors.New("short line")
	}

	var r Record

	r.Date = line[0]
	r.Nins, err = strconv.ParseInt(line[1], 0, 64)
	if err != nil {
		return nil, err
	}
	r.Ndel, err = strconv.ParseInt(line[2], 0, 64)
	if err != nil {
		return nil, err
	}
	r.Commit = line[3]
	r.Author = line[4]
	r.File = line[5]

	return &r, nil
}

func chart_draw_pchg(path string, nloc []Nloc, stime time.Time, etime time.Time, span int, width vg.Length, height vg.Length, codebase string) error {
	var v plotter.Values

	st := time.Unix(0, 0)
	et := st.AddDate(0, 0, span)

	sn := float64(nloc[0].Nloc)

	var now time.Time
	var pinc float64
	for i := 0; i < len(nloc); i++ {
		now = nloc[i].Date
		if now.Before(et) {
			continue
		}

		cn := float64(nloc[i].Nloc)

		if sn < cn {
			pinc = 100 - (sn / cn) * 100.0
		} else if (sn > cn) {
			pinc = 100 - (cn / sn) * 100.0
		} else {
			pinc = 0.0
		}

		sn = cn
		st = now
		et = st.AddDate(0, 0, span)

		if now.Before(stime) || now.After(etime) {
			continue
		}

		v = append(v, pinc)
	}

	w := vg.Points(2)

	bc, err := plotter.NewBarChart(v, w)
	if err != nil {
		return err
	}

	bc.LineStyle.Width = vg.Length(0)
	bc.Color = plotutil.Color(0)
	bc.Offset = -w

	p, err := plot.New()
	if err != nil {
		return err
	}

	if codebase != "" {
		p.Title.Text = fmt.Sprintf("Percent change in %s per %d day period", codebase, span)
	} else {
		p.Title.Text = fmt.Sprintf("Percent change per %d day period", span)
	}
	p.X.Label.Text = fmt.Sprintf("%d day periods starting on %v", span, stime)
	p.Y.Label.Text = "Percentage Change (%)"

	p.Add(bc)
	p.Add(plotter.NewGrid())

	return p.Save(width, height, path)
}

func chart_draw_nloc(path string, nloc []Nloc, stime time.Time, etime time.Time, width vg.Length, height vg.Length, codebase string) error {
	var pts plotter.XYs
	for i := 0; i < len(nloc); i++ {
		now := nloc[i].Date

		if now.Before(stime) || now.After(etime) {
			continue
		}
		pt := plotter.XY{float64(i), float64(nloc[i].Nloc)}
		pts = append(pts, pt)
	}

	lc, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	lc.LineStyle.Width = vg.Points(1)

	p, err := plot.New()
	if err != nil {
		return err
	}

	if codebase != "" {
		p.Title.Text = fmt.Sprintf("Total number of lines of code in %s ", codebase)
	} else {
		p.Title.Text = "Total number of lines of code per day"
	}
	p.X.Label.Text = fmt.Sprintf("Days since %v", stime)
	p.Y.Label.Text = "Number of lines of code"

	p.Add(lc)
	p.Add(plotter.NewGrid())

	return p.Save(width, height, path)
}

func main() {
	var inpath string
	var nlocfile string
	var pcntfile string
	var codebase string
	var pcntspan int
	var awidth float64
	var aheight float64
	var width vg.Length
	var height vg.Length
	var firstday int
	var lastday int

	flag.StringVar(&inpath, "infile", "", "Input path to .csv")
	flag.StringVar(&nlocfile, "nloc", "", "Output path to num lines of code over time chart (png, pdf, svg, etc)")
	flag.StringVar(&codebase, "codebase", "", "Name of codebase for chart title")
	flag.StringVar(&pcntfile, "pcnt", "", "Output path to %change over time chart (png, pdf, svg, etc)")
	flag.IntVar(&pcntspan, "pspan", 7, "Number of days per data point in the %change chart")
	flag.Float64Var(&awidth, "width", 10.0, "Chart width (inches)")
	flag.Float64Var(&aheight, "height", 7.5, "Chart height (inches)")
	flag.IntVar(&firstday, "firstday", 0, "Days since the beginning of the data to start chart on")
	flag.IntVar(&lastday, "lastday", -1, "Days since the beginning of the data to end the chart on (-1 = end of data)")
	flag.Parse()

	width = vg.Length(awidth)
	height = vg.Length(height)

	if inpath == "" {
		panic(errors.New("infile required"))
	}

	infile, err := reader_setup(inpath)
	if err != nil {
		panic(err)
	}

	if firstday < 0 {
		panic(errors.New("first day can't start before the data"));
	}

	if lastday != -1 && lastday < firstday {
		panic(errors.New("last day comes before first day"))
	}

	var totnloc int64 = 0
	var curdate time.Time
	var newdate time.Time
	var curnloc int64 = 0
	var nloc []Nloc
	first := true
	for {
		r, err := record_get(infile)
		if err != nil {
			panic(err)
		}
		if r == nil {
			break
		}

		if first { // first record, set start date
			curdate, err = parsedate(r.Date)
			if err != nil {
				panic(err)
			}
			first = false
		}

		newdate, err = parsedate(r.Date)
		if err != nil {
			panic(err)
		}

		if newdate.Before(curdate) {
			panic(errors.New(fmt.Sprintf("date %s before %s", newdate.String(), curdate.String())))
		}

		if newdate.Equal(curdate) {
			curnloc += r.Nins - r.Ndel // accumulate to have one data point per day
			continue
		}

		for curdate.Before(newdate) { // days without data are flat
			n := Nloc{curdate, totnloc}
			nloc = append(nloc, n)
			curdate = curdate.AddDate(0, 0, 1)
		}

		totnloc += curnloc
		curnloc = r.Nins - r.Ndel
		curdate = newdate
	}

	if curnloc != 0 {
		totnloc += curnloc

		n := Nloc{curdate, totnloc}
		nloc = append(nloc, n)

		curnloc = 0
	}

	if len(nloc) == 0 {
		fmt.Printf("WARNING: No records, exiting\n")
		return
	}

	stime := nloc[0].Date.AddDate(0, 0, firstday)

	var etime time.Time
	if lastday == -1 {
		etime = nloc[len(nloc)-1].Date
	} else {
		etime = nloc[0].Date.AddDate(0, 0, lastday)
	}

	fmt.Printf("Total Days: %d\n", len(nloc))
	fmt.Printf("Total NLOC %v %v through %v\n", totnloc, nloc[0].Date, nloc[len(nloc)-1].Date)
	fmt.Printf("Charts start on %v and end on %v\n", stime, etime)

	if nlocfile != "" {
		err = chart_draw_nloc(nlocfile, nloc, stime, etime, width, height, codebase)
		if err != nil {
			panic(err)
		}
	}

	if pcntfile != "" {
		err = chart_draw_pchg(pcntfile, nloc, stime, etime, pcntspan, width, height, codebase)
		if err != nil {
			panic(err)
		}
	}
}
