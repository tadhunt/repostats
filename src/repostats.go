package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg"
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

func reader_setup() (*csv.Reader, error) {
	file, err := os.Open("commitdata.csv")
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)

	return reader, nil
}

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

func chart_draw_pchg(nloc []Nloc) error {
	var v plotter.Values

	for i := 7; i < len(nloc); i += 7 {
		start := float64(nloc[i-7].Nloc)
		cur := float64(nloc[i].Nloc)

		var pinc float64
		if start < cur {
			pinc = 100 - (start/cur)*100.0
		} else {
			pinc = 100 - (cur/start)*100.0
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

	p.Add(bc)

	return p.Save(10, 5, "pcnt-chg-over-time.pdf")
}

func chart_draw_nloc(nloc []Nloc) error {
	pts := make(plotter.XYs, len(nloc))
	for i := 0; i < len(nloc); i++ {
		pt := &pts[i]
		pt.X = float64(i)
		pt.Y = float64(nloc[i].Nloc)
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

	p.Add(lc)

	return p.Save(10, 5, "nloc-over-time.pdf")
}

func main() {
	infile, err := reader_setup()
	if err != nil {
		panic(err)
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

	fmt.Printf("Total Days: %d\n", len(nloc))

	fmt.Printf("Total NLOC %v %v through %v\n", totnloc, nloc[0].Date, nloc[len(nloc)-1].Date)

	err = chart_draw_nloc(nloc)
	if err != nil {
		panic(err)
	}

	err = chart_draw_pchg(nloc)
	if err != nil {
		panic(err)
	}
}
