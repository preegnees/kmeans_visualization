package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"github.com/trimmer-io/go-csv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type data struct {
	X      float64
	Y      float64
	Points [][]float64
}

var myColor = color.RGBA{
	R: uint8(rand.Int31n(255)),
	G: uint8(rand.Int31n(255)),
	B: uint8(rand.Int31n(255)),
	A: 255,
}

var (
	FOLDER    = ""
	PY_SCRIPT = ""
	DATA_PATH = ""
)

var (
	AMOUNT_OF_POINTS  = -1
	AMOUNT_OF_CENTERS = -1
	THRESHOLD         = -1.0
	COUNT_ITER        = -1
	DEBUG             = 0
)

type istorage interface {
	getData() clusters.Observations
}

type storage struct {
	IsData bool
}

func (s *storage) getData() clusters.Observations {

	var d clusters.Observations
	if s.IsData {
		type Demo struct {
			Len1 string `csv:"len1"`
			Wid1 string `csv:"wid1"`
			// Len2  string `csv:"len2"`
			// Wid2  string `csv:"wid2"`
			// Class string `csv:"cl"`
		}
		type FrameSequence []*Demo

		fn := func(path string) (FrameSequence, error) {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			seq := make(FrameSequence, 0)
			if err := csv.Unmarshal('	', b, &seq); err != nil {
				return nil, err
			}
			return seq, nil
		}

		fr, err := fn(DATA_PATH)
		if err != nil {
			panic(err)
		}

		for _, v := range fr {
			// fmt.Printf("idx: %d, val: %v", i, v)
			x, err := strconv.ParseFloat(strings.Replace(v.Len1, ",", ".", -1), 64)
			y, err := strconv.ParseFloat(strings.Replace(v.Wid1, ",", ".", -1), 64)
			if err != nil {
				panic(err)
			}
			d = append(d, clusters.Coordinates{
				x,
				y,
			})
		}
		return d
	} else {
		for x := 0; x < AMOUNT_OF_POINTS; x++ {
			d = append(d, clusters.Coordinates{
				rand.Float64(),
				rand.Float64(),
			})
		}
	}

	if DEBUG == 1 {
		os.Exit(0)
	}

	return d
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	FOLDER = os.Getenv("IMG_FOLDER")
	DATA_PATH = os.Getenv("DATA_PATH")
	PY_SCRIPT = os.Getenv("PY_SCRIPT")
	AMOUNT_OF_POINTS, err = strconv.Atoi(os.Getenv("AMOUNT_OF_POINTS"))
	AMOUNT_OF_CENTERS, err = strconv.Atoi(os.Getenv("AMOUNT_OF_CENTERS"))
	THRESHOLD, err = strconv.ParseFloat(os.Getenv("THRESHOLD"), 64)
	COUNT_ITER, err = strconv.Atoi(os.Getenv("COUNT_ITER"))
	DEBUG, err = strconv.Atoi(os.Getenv("DEBUG"))

	var isData bool = false

	if DATA_PATH != "" {
		_, err := os.Open(DATA_PATH)
		if err != nil {
			panic(err)
		}
		isData = true
	}

	if err != nil {
		panic(err)
	}

	if err := clear(); err != nil {
		panic(err)
	}

	if err := km(&storage{IsData: isData}); err != nil {
		panic(err)
	}

	if err := show(); err != nil {
		panic(err)
	}
}

func clear() error {

	err := os.RemoveAll(FOLDER)
	if err != nil {
		return err
	}
	err = os.MkdirAll(FOLDER, 0777)
	if err != nil {
		return err
	}
	return nil
}

func show() error {

	cmd := exec.Command("python", PY_SCRIPT)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	copyOutput := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}
	go copyOutput(stdout)
	go copyOutput(stderr)
	cmd.Wait()

	return nil
}

func km(st istorage) error {

	km, err := kmeans.New(MyPlot{}, THRESHOLD, COUNT_ITER)
	if err != nil {
		return err
	}

	clusters, err := km.Partition(st.getData(), AMOUNT_OF_CENTERS)
	if err != nil {
		return err
	}

	var dt []data = make([]data, AMOUNT_OF_CENTERS)

	for _, c := range clusters {
		ps := make([][]float64, 0)
		for _, v := range c.Observations {
			ps = append(ps, []float64{float64(v.Coordinates()[0]), float64(v.Coordinates()[1])})
		}
		dt = append(dt, data{
			X:      c.Center[0],
			Y:      c.Center[1],
			Points: ps,
		})
	}

	return nil
}

type MyPlot struct{}

func (m MyPlot) Plot(cc clusters.Clusters, iteration int) error {

	rand.Seed(int64(0))
	p := plot.New()
	p.Title.Text = "Points Example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	p.Add(plotter.NewGrid())

	for _, c := range cc {
		ps := make([][]float64, 0)
		for _, v := range c.Observations {
			ps = append(ps, []float64{float64(v.Coordinates()[0]), float64(v.Coordinates()[1])})
		}
		d := data{
			X:      c.Center[0],
			Y:      c.Center[1],
			Points: ps,
		}

		scatterPoints := addPoint(d)
		sp, err := plotter.NewScatter(scatterPoints)
		if err != nil {
			return err
		}

		sp.GlyphStyle.Color = color.RGBA{
			R: uint8(rand.Int31n(255)),
			G: uint8(rand.Int31n(255)),
			B: uint8(rand.Int31n(255)),
			A: 255,
		}
		p.Add(sp)

		scatterCenters := addCenters(d)
		sC, err := plotter.NewScatter(scatterCenters)
		if err != nil {
			return err
		}
		sC.GlyphStyle.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}
		sC.GlyphStyle.Radius = 7
		p.Add(sC)
	}

	path := fmt.Sprintf("%s\\points_%d.png", FOLDER, iteration)
	if err := p.Save(7*vg.Inch, 6*vg.Inch, path); err != nil {
		panic(err)
	}

	return nil
}

func addPoint(dt data) plotter.XYs {

	pts := make(plotter.XYs, 0)
	for _, p := range dt.Points {
		pts = append(pts, plotter.XY{
			X: p[0],
			Y: p[1],
		})
	}

	return pts
}

func addCenters(dt data) plotter.XYs {

	pts := make(plotter.XYs, 0)
	pts = append(pts, plotter.XY{
		X: dt.X,
		Y: dt.Y,
	})

	return pts
}
